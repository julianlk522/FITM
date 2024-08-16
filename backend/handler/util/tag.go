package handler

import (
	"database/sql"
	"fmt"
	"math"
	"slices"
	"strings"

	"oitm/db"
	"oitm/model"
	"oitm/query"

	"net/http"

	"github.com/go-chi/render"
)

// Get tags for link
func ScanTagPageLink(link_sql *query.TagPageLink) (*model.LinkSignedIn, error) {
	var l = &model.LinkSignedIn{}

	err := db.Client.
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

func GetUserTagForLink(login_name string, link_id string) (*model.Tag, error) {
	var id, cats, last_updated sql.NullString

	err := db.Client.
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

func ScanTagRankings(tag_rankings_sql *query.TagRankings) (*[]model.TagRankingPublic, error) {
	rows, err := db.Client.Query(tag_rankings_sql.Text)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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



// Get top global cats
func ScanAndSplitGlobalCats(global_cats_sql *query.GlobalCats) (*[]string, error) {
	rows, err := db.Client.Query(global_cats_sql.Text)
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

func GetCatCounts(categories *[]string) (*[]model.CatCount, error) {
	num_cats := len(*categories)
	var category_counts []model.CatCount = make([]model.CatCount, num_cats)

	for i := 0; i < num_cats; i++ {
		category_counts[i].Category = (*categories)[i]

		cat_count_sql := fmt.Sprintf(`SELECT count(*) as count_with_cat FROM (%s);`, query.NewLinkIDs((*categories)[i]).Text)

		var c sql.NullInt32
		err := db.Client.QueryRow(cat_count_sql).Scan(&c)
		if err != nil {
			return nil, err
		}

		category_counts[i].Count = c.Int32
	}

	SortAndLimitCatCounts(&category_counts)

	return &category_counts, nil

}

func GetCatCountsDuringPeriod(categories *[]string, period string) (*[]model.CatCount, error) {
	num_cats := len(*categories)
	var category_counts []model.CatCount = make([]model.CatCount, num_cats)

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
		err := db.Client.QueryRow(cat_count_sql).Scan(&c)
		if err != nil {
			return nil, err
		}

		category_counts[i].Count = c.Int32
	}

	SortAndLimitCatCounts(&category_counts)

	return &category_counts, nil

}

func RenderCategoryCounts(category_counts *[]model.CatCount, w http.ResponseWriter, r *http.Request, ) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, category_counts)
}



// Add tag
func UserHasTaggedLink(login_name string, link_id string) (bool, error) {
	var t sql.NullString
	err := db.Client.QueryRow("SELECT id FROM Tags WHERE submitted_by = ? AND link_id = ?;", login_name, link_id).Scan(&t)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil

}

func AssignNewTagIDToRequest(res sql.Result, request *model.NewTagRequest) error {
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	request.ID = id
	return nil
}



// Edit tag
func UserSubmittedTagWithID(login_name string, tag_id string) (bool, error) {
	var t sql.NullString
	err := db.Client.QueryRow("SELECT id FROM Tags WHERE submitted_by = ? AND id = ?;", login_name, tag_id).Scan(&t)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil

}

func AlphabetizeCats(cats string) string {
	split_categories := strings.Split(cats, ",")
	slices.Sort(split_categories)

	return strings.Join(split_categories, ",")
}

func GetLinkIDFromTagID(tag_id string) (string, error) {
	var link_id sql.NullString
	err := db.Client.QueryRow("SELECT link_id FROM Tags WHERE id = ?;", tag_id).Scan(&link_id)
	if err != nil {
		return "", err
	}

	return link_id.String, nil
}



// Calculate global cats
func CalculateAndSetGlobalCats(link_id string) error {
	overlap_scores_sql := query.NewTopOverlapScores(link_id)
	if overlap_scores_sql.Error != nil {
		return overlap_scores_sql.Error
	}

	rows, err := db.Client.Query(overlap_scores_sql.Text)
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
	alphabetized_cats := AlphabetizeOverlapScoreCats(overlap_scores)
	
	
	// Assign to global cats if >= 50% of max category score
	var global_cats string
	for _, cat := range alphabetized_cats {
		if overlap_scores[cat] >= max_cat_score * 0.5 {
			global_cats += cat + ","
		}
	}

	// Remove trailing comma
	if len(global_cats) > 0 && strings.HasSuffix(global_cats, ",") {
		global_cats = global_cats[:len(global_cats)-1]
	}

	err = SetGlobalCats(link_id, global_cats)
	if err != nil {
		return err
	}
	
	return nil
}

func AlphabetizeOverlapScoreCats(scores map[string]float32) []string {
	cats := make([]string, 0, len(scores))
	for cat := range scores {
		cats = append(cats, cat)
	}
	slices.Sort(cats)

	return cats
}


func SetGlobalCats(link_id string, text string) error {
	_, err := db.Client.Exec(`
		UPDATE Links 
		SET global_cats = ? 
		WHERE id = ?`, 
	text, 
	link_id)
	if err != nil {
		return err
	}

	return nil
}