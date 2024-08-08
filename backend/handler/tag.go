package handler

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"golang.org/x/exp/slices"

	query "oitm/db/query"
	e "oitm/error"
	m "oitm/middleware"
	"oitm/model"
)

func GetTagsForLink(w http.ResponseWriter, r *http.Request) {
	link_id := chi.URLParam(r, "link_id")
	if link_id == "" {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoLinkID))
		return
	}

	link_exists, err := _LinkExists(link_id)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}
	if !link_exists {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoLinkWithID))
		return
	}

	req_user_id := r.Context().Value(m.UserIDKey).(string)
	link_sql := query.NewTagPageLink(link_id, req_user_id)
	if link_sql.Error != nil {
		render.Render(w, r, e.ErrInvalidRequest(link_sql.Error))
		return
	}

	link, err := _ScanTagPageLink(link_sql)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	req_login_name := r.Context().Value(m.LoginNameKey).(string)
	user_tag, err := _GetUserTagForLink(req_login_name, link_id)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	tag_rankings_sql := query.NewTagRankingsForLink(link_id)
	if tag_rankings_sql.Error != nil {
		render.Render(w, r, e.ErrInvalidRequest(tag_rankings_sql.Error))
		return
	}

	tag_rankings, err := _ScanTagRankings(tag_rankings_sql)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	tag_page := model.TagPage{
		Link: link,
		UserTag: user_tag,
		TagRankings: tag_rankings,
	}
	render.JSON(w, r, tag_page)
}

func _ScanTagPageLink(link_sql *query.TagPageLink) (*model.LinkSignedIn, error) {
	var l = &model.LinkSignedIn{}

	err := DBClient.
		QueryRow(link_sql.Text).
		Scan(
			&l.ID, 
			&l.URL, 
			&l.SubmittedBy, 
			&l.SubmitDate, 
			&l.Categories, 
			&l.Summary, 
			&l.SummaryCount,
			&l.LikeCount, 
			&l.ImgURL, 
			&l.IsLiked, 
			&l.IsCopied,
		)
	if err != nil {
		return nil, err
	}

	return l, nil
}

func _GetUserTagForLink(login_name string, link_id string) (*model.Tag, error) {
	var id, cats, last_updated sql.NullString

	err := DBClient.
		QueryRow("SELECT id, categories, last_updated FROM 'Tags' WHERE submitted_by = ? AND link_id = ?;", login_name, link_id).
		Scan(&id, &cats, &last_updated)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &model.Tag{
		ID: id.String,
		Categories: cats.String,
		LastUpdated: last_updated.String,
		LinkID: link_id,
		SubmittedBy: login_name,
	}, nil
}

func _ScanTagRankings(tag_rankings_sql *query.TagRankings) (*[]model.TagRankingPublic, error) {
	rows, err := DBClient.Query(tag_rankings_sql.Text)
	if err != nil {
		return nil, err
	}

	tag_rankings := []model.TagRankingPublic{}

	for rows.Next() {
		var tag model.TagRankingPublic
		err = rows.Scan(
			&tag.LifeSpanOverlap, 
			&tag.Categories, 
			&tag.SubmittedBy, 
			&tag.LastUpdated,
		)
		if err != nil {
			return nil, err
		}
		tag_rankings = append(tag_rankings, tag)
	}

	return &tag_rankings, nil
}

// TOP GLOBAL CATS (MOST-USED OVERALL OR IN PERIOD)
func GetTopGlobalCats(w http.ResponseWriter, r *http.Request) {	
	global_cats_sql := query.NewAllGlobalCats()

	period_params := r.URL.Query().Get("period")
	has_period := period_params != ""
	if has_period {
		global_cats_sql = global_cats_sql.DuringPeriod(period_params)
	}

	if global_cats_sql.Error != nil {
		render.Render(w, r, e.ErrInvalidRequest(global_cats_sql.Error))
		return
	}

	cats, err := _ScanGlobalCategories(global_cats_sql)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	var counts *[]model.CategoryCount
	if has_period {
		counts, err = _GetCategoryCountsDuringPeriod(cats, period_params)
	} else {
		counts, err = _GetCategoryCounts(cats)
	}

	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}
	_RenderCategoryCounts(counts, w, r)
}

func _ScanGlobalCategories(global_cats_sql *query.GlobalCats) (*[]string, error) {
	rows, err := DBClient.Query(global_cats_sql.Text)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var categories []string

	for rows.Next() {
		var cats_field string
		err = rows.Scan(&cats_field)
		if err != nil {
			return nil, err
		}
		
		cats_field = strings.ToLower(cats_field)
		
		// Split each row into individual categories if multiple
		if strings.Contains(cats_field, ",") {
			split_cats := strings.Split(cats_field, ",")

			for _, cat := range(split_cats) {
				if !slices.Contains(categories, cat) {
					categories = append(categories, cat)
				}
			}

		// Single row category
		} else {
			if !slices.Contains(categories, cats_field) {
				categories = append(categories, cats_field)
			}
		}
	}

	return &categories, nil
}

func _GetCategoryCounts(categories *[]string) (*[]model.CategoryCount, error) {
	num_cats := len(*categories)
	var category_counts []model.CategoryCount = make([]model.CategoryCount, num_cats)

	for i := 0; i < num_cats; i++ {
		category_counts[i].Category = (*categories)[i]

		cat_count_sql := fmt.Sprintf(`SELECT count(*) as count_with_cat FROM (%s);`, query.NewLinkIDs((*categories)[i]).Text)

		var c sql.NullInt32
		err := DBClient.QueryRow(cat_count_sql).Scan(&c)
		if err != nil {
			return nil, err
		}

		category_counts[i].Count = c.Int32
	}

	_SortAndLimitCategoryCounts(&category_counts)

	return &category_counts, nil

}

func _GetCategoryCountsDuringPeriod(categories *[]string, period string) (*[]model.CategoryCount, error) {
	num_cats := len(*categories)
	var category_counts []model.CategoryCount = make([]model.CategoryCount, num_cats)

	period_clause, err := query.GetPeriodClause(period)
	if err != nil {
		return nil, err
	}

	for i := 0; i < num_cats; i++ {
		category_counts[i].Category = (*categories)[i]

		cat_count_sql := fmt.Sprintf(
			`SELECT count(*) as count_with_cat FROM (%s) WHERE %s;`, 
			query.NewLinkIDs((*categories)[i]).Text, 
			period_clause,
		)
		cat_count_sql = strings.Replace(cat_count_sql, "SELECT id", "SELECT id, submit_date", 1)

		var c sql.NullInt32
		err := DBClient.QueryRow(cat_count_sql).Scan(&c)
		if err != nil {
			return nil, err
		}

		category_counts[i].Count = c.Int32
	}

	_SortAndLimitCategoryCounts(&category_counts)

	return &category_counts, nil

}

func _RenderCategoryCounts(category_counts *[]model.CategoryCount, w http.ResponseWriter, r *http.Request, ) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, category_counts)
}

// ADD NEW TAG
func AddTag(w http.ResponseWriter, r *http.Request) {
	tag_data := &model.NewTagRequest{}
	if err := render.Bind(r, tag_data); err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}
	
	link_exists, err := _LinkExists(tag_data.LinkID)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	} else if !link_exists {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoLinkWithID))
		return
	}

	req_login_name := r.Context().Value(m.LoginNameKey).(string)
	duplicate, err := _UserHasSubmittedTagToLink(req_login_name, tag_data.LinkID)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	} else if duplicate {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrDuplicateTag))
		return
	}
	
	tag_data.Categories = strings.ToLower(tag_data.Categories)
	res, err := DBClient.Exec(
		"INSERT INTO Tags VALUES(?,?,?,?,?);", 
		nil, 
		tag_data.LinkID, 
		tag_data.Categories, 
		req_login_name, 
		tag_data.LastUpdated,
	)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	if err = _RecalculateGlobalCategories(tag_data.LinkID); err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	if err := _AssignNewTagIDToRequest(res, tag_data); err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, tag_data)
}

func _UserHasSubmittedTagToLink(login_name string, link_id string) (bool, error) {
	var t sql.NullString
	err := DBClient.QueryRow("SELECT id FROM Tags WHERE submitted_by = ? AND link_id = ?;", login_name, link_id).Scan(&t)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil

}

func _AssignNewTagIDToRequest(res sql.Result, request *model.NewTagRequest) error {
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	request.ID = id
	return nil
}

// EDIT TAG
func EditTag(w http.ResponseWriter, r *http.Request) {
	edit_tag_data := &model.EditTagRequest{}
	if err := render.Bind(r, edit_tag_data); err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	req_login_name := r.Context().Value(m.LoginNameKey).(string)
	owns_tag, err := _UserHasSubmittedTag(req_login_name, edit_tag_data.ID)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoTagWithID))
		return
	} else if !owns_tag {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrDoesntOwnTag))
		return
	}

	edit_tag_data.Categories = _AlphabetizeCategories(edit_tag_data.Categories)

	_, err = DBClient.Exec(
		`UPDATE Tags 
		SET categories = ?, last_updated = ? 
		WHERE id = ?;`, 
		edit_tag_data.Categories, 
		edit_tag_data.LastUpdated, 
		edit_tag_data.ID,
	)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	link_id, err := _GetLinkIDFromTagID(edit_tag_data.ID)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	} else if err = _RecalculateGlobalCategories(link_id); err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, edit_tag_data)

}

func _UserHasSubmittedTag(login_name string, tag_id string) (bool, error) {
	var t sql.NullString
	err := DBClient.QueryRow("SELECT id FROM Tags WHERE submitted_by = ? AND id = ?;", login_name, tag_id).Scan(&t)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil

}

func _AlphabetizeCategories(categories string) string {
	split_categories := strings.Split(categories, ",")
	slices.Sort(split_categories)

	return strings.Join(split_categories, ",")
}

func _GetLinkIDFromTagID(tag_id string) (string, error) {
	var link_id sql.NullString
	err := DBClient.QueryRow("SELECT link_id FROM Tags WHERE id = ?;", tag_id).Scan(&link_id)
	if err != nil {
		return "", err
	}

	return link_id.String, nil
}

// Recalculate global cats for a link whose tags changed
func _RecalculateGlobalCategories(link_id string) error {
	overlap_scores_sql := query.
		NewTopOverlapScores(link_id).
		Limit(TOP_OVERLAP_SCORES_LIMIT)
	if overlap_scores_sql.Error != nil {
		return overlap_scores_sql.Error
	}

	rows, err := DBClient.Query(overlap_scores_sql.Text)
	if err != nil {
		return err
	}
	defer rows.Close()

	tag_rankings := []model.TagRanking{}
	for rows.Next() {
		var t model.TagRanking
		err = rows.Scan(&t.LifeSpanOverlap, &t.Categories)
		if err != nil {
			return err
		}
		tag_rankings = append(tag_rankings, t)
	}

	overlap_scores := make(map[string]float32)
	var max_cat_score float32

	for _, tag := range tag_rankings {
		
		// square root lifespan overlap to smooth out scores
		// (allows brand-new tags to still have some influence)
		tag.LifeSpanOverlap = float32(math.Sqrt(float64(tag.LifeSpanOverlap)))
		
		cat_field_lc := strings.ToLower(tag.Categories)

		// multiple categories
		if strings.Contains(cat_field_lc, ",") {
			cats := strings.Split(cat_field_lc, ",")
			for _, cat := range cats {
				overlap_scores[cat] += tag.LifeSpanOverlap

				if overlap_scores[cat] > max_cat_score {
					max_cat_score = overlap_scores[cat]
				}
			}

		// single category
		} else {
			overlap_scores[cat_field_lc] += tag.LifeSpanOverlap

			if overlap_scores[cat_field_lc] > max_cat_score {
				max_cat_score = overlap_scores[cat_field_lc]
			}
		}
	}

	// Alphabetize so global categories are assigned in order
	alphabetized_cats := _AlphabetizeOverlapScoreCategories(overlap_scores)
	
	var global_cats string

	// Assign to global cats if >= 50% of max category score
	for _, cat := range alphabetized_cats {
		if overlap_scores[cat] >= max_cat_score * 0.5 {
			global_cats += cat + ","
		}
	}

	if len(global_cats) > 0 && strings.HasSuffix(global_cats, ",") {
		global_cats = global_cats[:len(global_cats)-1]
	}

	_, err = DBClient.Exec("UPDATE Links SET global_cats = ? WHERE id = ?;", global_cats, link_id)
	if err != nil {
		return err

	}

	return nil
}

func _AlphabetizeOverlapScoreCategories(scores map[string]float32) []string {
	cats := make([]string, 0, len(scores))
	for cat := range scores {
		cats = append(cats, cat)
	}
	slices.Sort(cats)

	return cats
}