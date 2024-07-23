package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"golang.org/x/exp/slices"

	query "oitm/db/query"
	"oitm/model"
)

func GetTagsForLink(w http.ResponseWriter, r *http.Request) {
	link_id := chi.URLParam(r, "link_id")
	if link_id == "" {
		render.Render(w, r, ErrInvalidRequest(ErrNoLinkID))
		return
	}

	link_exists, err := _LinkExists(link_id)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if !link_exists {
		render.Render(w, r, ErrInvalidRequest(ErrNoLinkWithID))
		return
	}

	req_user_id, req_login_name, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	get_link_sql := query.NewGetTagPageLink(link_id, req_user_id)
	if get_link_sql.Error != nil {
		render.Render(w, r, ErrInvalidRequest(get_link_sql.Error))
		return
	}

	link, err := _ScanTagPageLink(get_link_sql)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	user_tag, err := _GetUserTagForLink(req_login_name, link_id)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	earliest_tags_sql := query.NewGetEarliestTags(link_id)
	if earliest_tags_sql.Error != nil {
		render.Render(w, r, ErrInvalidRequest(earliest_tags_sql.Error))
		return
	}

	earliest_tags, err := _ScanEarliestTags(earliest_tags_sql)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	tag_page := model.TagPage{
		Link: link,
		UserTag: user_tag,
		EarliestTags: earliest_tags,
	}
	render.JSON(w, r, tag_page)
}

func _ScanTagPageLink(get_link_sql *query.GetTagPageLink) (*model.LinkSignedIn, error) {
	var link = &model.LinkSignedIn{}

	err := DBClient.QueryRow(get_link_sql.Text).Scan(&link.ID, &link.URL, &link.SubmittedBy, &link.SubmitDate, &link.Categories, &link.Summary, &link.LikeCount, &link.ImgURL, &link.IsLiked, &link.IsCopied)
	if err != nil {
		return nil, err
	}

	return link, nil
}

func _GetUserTagForLink(login_name string, link_id string) (*model.Tag, error) {
	var id, cats, last_updated sql.NullString

	err := DBClient.QueryRow("SELECT id, categories, last_updated FROM 'Tags' WHERE submitted_by = ? AND link_id = ?;", login_name, link_id).Scan(&id, &cats, &last_updated)
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

func _ScanEarliestTags(earliest_cats_sql *query.GetEarliestTags) (*[]model.EarlyTagPublic, error) {
	rows, err := DBClient.Query(earliest_cats_sql.Text)
	if err != nil {
		return nil, err
	}

	earliest_tags := []model.EarlyTagPublic{}

	for rows.Next() {
		var tag model.EarlyTagPublic
		err = rows.Scan(&tag.LifeSpanOverlap, &tag.Categories, &tag.SubmittedBy, &tag.LastUpdated)
		if err != nil {
			return nil, err
		}
		earliest_tags = append(earliest_tags, tag)
	}

	return &earliest_tags, nil
}

// GET MOST-USED TAG CATEGORIES
func GetTopTagCategories(w http.ResponseWriter, r *http.Request) {	
	get_global_cats_sql := query.NewGetAllGlobalCategories()
	if get_global_cats_sql.Error != nil {
		render.Render(w, r, ErrInvalidRequest(get_global_cats_sql.Error))
		return
	}

	categories, err := _ScanGlobalCategories(get_global_cats_sql)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	counts, err := _GetCategoryCounts(categories)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	_RenderCategoryCounts(counts, w, r)
}

func GetTopTagCategoriesByPeriod(w http.ResponseWriter, r *http.Request) {
	period_params := chi.URLParam(r, "period")
	if period_params == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("no period provided")))
		return
	}

	get_global_cats_sql := query.NewGetAllGlobalCategories().FromPeriod(period_params)
	if get_global_cats_sql.Error != nil {
		render.Render(w, r, ErrInvalidRequest(get_global_cats_sql.Error))
		return
	}

	categories, err := _ScanGlobalCategories(get_global_cats_sql)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	counts, err := _GetCategoryCountsDuringPeriod(categories, period_params)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	_RenderCategoryCounts(counts, w, r)
}

func _ScanGlobalCategories(get_global_cats_sql *query.GetAllGlobalCategories) (*[]string, error) {
	rows, err := DBClient.Query(get_global_cats_sql.Text)
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

		get_cat_count_sql := fmt.Sprintf(`SELECT count(*) as count_with_cat FROM (%s);`, query.NewGetLinkIDs((*categories)[i]).Text)

		var c sql.NullInt32
		err := DBClient.QueryRow(get_cat_count_sql).Scan(&c)
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

		get_cat_count_sql := fmt.Sprintf(`SELECT count(*) as count_with_cat FROM (%s) WHERE %s;`, query.NewGetLinkIDs((*categories)[i]).Text, period_clause)
		get_cat_count_sql = strings.Replace(get_cat_count_sql, "SELECT id", "SELECT id, submit_date", 1)

		var c sql.NullInt32
		err := DBClient.QueryRow(get_cat_count_sql).Scan(&c)
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
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if strings.Count(tag_data.Categories, ",") > NEW_TAG_CATEGORY_LIMIT {
		render.Render(w, r, ErrInvalidRequest(ErrTooManyCategories))
		return
	}
	
	link_exists, err := _LinkExists(tag_data.LinkID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if !link_exists {
		render.Render(w, r, ErrInvalidRequest(ErrNoLinkWithID))
		return
	}

	_, req_login_name, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	duplicate, err := _UserHasSubmittedTagToLink(req_login_name, tag_data.LinkID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if duplicate {
		render.Render(w, r, ErrInvalidRequest(ErrDuplicateTag))
		return
	}
	
	tag_data.Categories = strings.ToLower(tag_data.Categories)
	res, err := DBClient.Exec("INSERT INTO Tags VALUES(?,?,?,?,?);", nil, tag_data.LinkID, tag_data.Categories, req_login_name, tag_data.LastUpdated)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if err = _RecalculateGlobalCategories(tag_data.LinkID); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if err := _AssignNewTagIDToRequest(res, tag_data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
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
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	_, req_login_name, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	
	owns_tag, err := _UserHasSubmittedTag(req_login_name, edit_tag_data.ID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(ErrNoTagWithID))
		return
	} else if !owns_tag {
		render.Render(w, r, ErrInvalidRequest(ErrDoesntOwnTag))
		return
	}

	edit_tag_data.Categories = _AlphabetizeCategories(edit_tag_data.Categories)

	_, err = DBClient.Exec(`UPDATE Tags 
	SET categories = ?, last_updated = ? 
	WHERE id = ?;`, 
	edit_tag_data.Categories, edit_tag_data.LastUpdated, edit_tag_data.ID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	link_id, err := _GetLinkIDFromTagID(edit_tag_data.ID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if err = _RecalculateGlobalCategories(link_id); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
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

// Recalculate global categories for a link whose tags changed
// (technically should affect all links that share 1+ categories but that's too complicated) 
// (many links will also not be seen enough to justify being updated constantly. makes enough sense to only update a link's global cats when a new tag is added to that link.)
func _RecalculateGlobalCategories(link_id string) error {
	get_overlap_scores_sql := query.NewGetTopOverlapScores(link_id).Limit(TOP_OVERLAP_SCORES_LIMIT)
	if get_overlap_scores_sql.Error != nil {
		return get_overlap_scores_sql.Error
	}

	rows, err := DBClient.Query(get_overlap_scores_sql.Text)
	if err != nil {
		return err
	}
	defer rows.Close()

	earliest_tags := []model.EarlyTag{}
	for rows.Next() {
		var t model.EarlyTag
		err = rows.Scan(&t.LifeSpanOverlap, &t.Categories)
		if err != nil {
			return err
		}
		earliest_tags = append(earliest_tags, t)
	}

	overlap_scores := make(map[string]float32)

	max_row_score := 1 / float32(len(earliest_tags))
	var max_score float32

	for _, tag := range earliest_tags {
		
		// square root lifespan overlap to smooth out scores
		// (allows brand-new tags to still have some influence)
		tag.LifeSpanOverlap = float32(math.Sqrt(float64(tag.LifeSpanOverlap)))
		
		cat_field_lc := strings.ToLower(tag.Categories)

		if strings.Contains(cat_field_lc, ",") {
			cats := strings.Split(cat_field_lc, ",")
			for _, cat := range cats {
				overlap_scores[cat] += tag.LifeSpanOverlap * max_row_score

				if overlap_scores[cat] > max_score {
					max_score = overlap_scores[cat]
				}
			}

		// single category
		} else {
			overlap_scores[cat_field_lc] += tag.LifeSpanOverlap * max_row_score

			if overlap_scores[cat_field_lc] > max_score {
				max_score = overlap_scores[cat_field_lc]
			}
		}
	}

	// Alphabetize so global categories are assigned in order
	alphabetized_cats := _AlphabetizeOverlapScoreCategories(overlap_scores)
	
	var global_cats string

	// Assign to global cats if >= 50% of max category score
	for _, cat := range alphabetized_cats {
		if overlap_scores[cat] >= max_score * 0.5 {
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