package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"golang.org/x/exp/slices"

	"oitm/model"
)

func GetTagsForLink(w http.ResponseWriter, r *http.Request) {
	link_id := chi.URLParam(r, "link_id")
	if link_id == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("no link id found")))
		return
	}

	// Check auth token
	var req_user_id string
	var req_login_name string
	claims, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if len(claims) > 0 {
		req_user_id = claims["user_id"].(string)
		req_login_name = claims["login_name"].(string)
	}

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check if link exists, Abort if invalid link ID provided
	var l sql.NullString
	err = db.QueryRow("SELECT id FROM Links WHERE id = ?;", link_id).Scan(&l)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("no link found with given ID")))
		return
	}

	// Get link
	get_link_sql := fmt.Sprintf(`SELECT links_id as link_id, url, submitted_by, submit_date, coalesce(categories,"") as categories, summary, COUNT('Link Likes'.id) as like_count, coalesce(is_liked,0) as is_liked, coalesce(is_copied,0) as is_copied, img_url
	FROM
		(
		SELECT id as links_id, url, submitted_by, submit_date, global_cats as categories, global_summary as summary, coalesce(img_url,"") as img_url
		FROM Links
		WHERE id = '%[1]s'
		)
	LEFT JOIN 'Link Likes'
	ON 'Link Likes'.link_id = links_id
	LEFT JOIN
		(
		SELECT id as like_id, count(*) as is_liked, user_id as luser_id, link_id as like_link_id2
		FROM 'Link Likes'
		WHERE luser_id = '%[2]s'
		GROUP BY like_id
		)
	ON like_link_id2 = link_id
	LEFT JOIN
		(
		SELECT id as copy_id, count(*) as is_copied, user_id as cuser_id, link_id as copy_link_id
		FROM 'Link Copies'
		WHERE cuser_id = '%[2]s'
		GROUP BY copy_id
		)
	ON copy_link_id = link_id`, link_id, req_user_id)
	var link model.LinkSignedIn
	err = db.QueryRow(get_link_sql).Scan(&link.ID, &link.URL, &link.SubmittedBy, &link.SubmitDate, &link.Categories, &link.Summary, &link.LikeCount, &link.IsLiked, &link.IsCopied, &link.ImgURL)
	if err != nil {
		log.Fatal(err)
	}

	// Get earliest tags and Lifespan-Overlap scores
	rows, err := db.Query(`SELECT (julianday('now') - julianday(last_updated)) / (julianday('now') - julianday(submit_date)) * 100 AS lifespan_overlap, categories, Tags.submitted_by, last_updated 
		FROM Tags 
		INNER JOIN Links 
		ON Links.id = Tags.link_id 
		WHERE link_id = ? 
		ORDER BY lifespan_overlap DESC 
		LIMIT 20;`, link_id)
	if err != nil {
		log.Fatal(err)
	}

	earliest_tags := []model.EarlyTagPublic{}
	for rows.Next() {
		var tag model.EarlyTagPublic
		err = rows.Scan(&tag.LifeSpanOverlap, &tag.Categories, &tag.SubmittedBy, &tag.LastUpdated)
		if err != nil {
			log.Fatal(err)
		}
		earliest_tags = append(earliest_tags, tag)
	}

	// Get user-submitted tag if user has submitted one
	var user_tag_id, user_tag_cats, user_tag_last_updated sql.NullString
	err = db.QueryRow("SELECT id, categories, last_updated FROM 'Tags' WHERE link_id = ? AND submitted_by = ?;", link_id, req_login_name).Scan(&user_tag_id, &user_tag_cats, &user_tag_last_updated)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Fatal(err)
		}

		tag_page := model.TagPage{
			Link: &link,
			UserTag: nil,
			TopTags: &earliest_tags,
		}
		render.JSON(w, r, tag_page)

	} else if user_tag_cats.Valid {
		user_tag := model.Tag{
			ID: user_tag_id.String,
			Categories: user_tag_cats.String,
			LastUpdated: user_tag_last_updated.String,
			LinkID: link_id,
			SubmittedBy: req_login_name,
		}

		tag_page := model.TagPage{
			Link: &link,
			UserTag: &user_tag,
			TopTags: &earliest_tags,
		}
		render.JSON(w, r, tag_page)
	}
}

// GET MOST-USED TAG CATEGORIES
func GetTopTagCategories(w http.ResponseWriter, r *http.Request) {

	// Limit 15 for now
	const LIMIT int = 15

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Get all global categories
	rows, err := db.Query(`SELECT global_cats
		FROM Links
		WHERE global_cats != ""
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Split global categories for each link into individual categories
	var categories []string
	for rows.Next() {
		var cat_field string
		err = rows.Scan(&cat_field)
		if err != nil {
			log.Fatal(err)
		}

		cat_field = strings.ToLower(cat_field)

		if strings.Contains(cat_field, ",") {
			split := strings.Split(cat_field, ",")

			for i := 0; i < len(split); i++ {
				if !slices.Contains(categories, split[i]) {
					categories = append(categories, split[i])
				}
			}
		} else {
			if !slices.Contains(categories, cat_field) {
				categories = append(categories, cat_field)
			}
		}
	}

	// get counts for each category
	var category_counts []model.CategoryCount = make([]model.CategoryCount, len(categories))
	for i := 0; i < len(categories); i++ {
		category_counts[i].Category = categories[i]

		get_cat_count_sql := fmt.Sprintf(`select count(*) as count_with_cat from (select id from Links where ',' || global_cats || ',' like '%%,%s,%%' group by id)`, categories[i])

		var c sql.NullInt32
		err = db.QueryRow(get_cat_count_sql).Scan(&c)
		if err != nil {
			log.Fatal(err)
		}

		category_counts[i].Count = c.Int32
	}

	slices.SortFunc(category_counts, model.SortCategories)

	// return top {LIMIT} categories and their counts
	render.Status(r, http.StatusOK)
	render.JSON(w, r, category_counts[0:LIMIT])
}

func GetTopTagCategoriesByPeriod(w http.ResponseWriter, r *http.Request) {

	// Limit 15 for now
	const LIMIT int = 15

	get_tag_cats_sql := `SELECT global_cats
		FROM Links
		WHERE global_cats != ""`

	switch chi.URLParam(r, "period") {
	case "day":
		get_tag_cats_sql += ` AND julianday('now') - julianday(submit_date) <= 2`
	case "week":
		get_tag_cats_sql += ` AND julianday('now') - julianday(submit_date) <= 8`
	case "month":
		get_tag_cats_sql += ` AND julianday('now') - julianday(submit_date) <= 31`
	case "year":
		get_tag_cats_sql += ` AND julianday('now') - julianday(submit_date) <= 366`
	default:
		render.Render(w, r, ErrInvalidRequest(errors.New("invalid period")))
		return
	}

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query(get_tag_cats_sql)
	if err != nil {
		log.Fatal(err)
	}

	// Split global categories for each link into individual categories
	var categories []string
	for rows.Next() {
		var cat_field string
		err = rows.Scan(&cat_field)
		if err != nil {
			log.Fatal(err)
		}

		cat_field = strings.ToLower(cat_field)

		if strings.Contains(cat_field, ",") {
			split := strings.Split(cat_field, ",")

			for i := 0; i < len(split); i++ {
				if !slices.Contains(categories, split[i]) {
					categories = append(categories, split[i])
				}
			}
		} else {
			if !slices.Contains(categories, cat_field) {
				categories = append(categories, cat_field)
			}
		}
	}

	// get counts for each category
	var category_counts []model.CategoryCount = make([]model.CategoryCount, len(categories))
	for i := 0; i < len(categories); i++ {
		category_counts[i].Category = categories[i]

		get_cat_count_sql := fmt.Sprintf(`select count(*) as count_with_cat from (select id, submit_date from Links where ',' || global_cats || ',' like '%%,%s,%%' group by id)`, categories[i])

		switch chi.URLParam(r, "period") {
			case "day":
				get_cat_count_sql += ` WHERE julianday('now') - julianday(submit_date) <= 2`
			case "week":
				get_cat_count_sql += ` WHERE julianday('now') - julianday(submit_date) <= 8`
			case "month":
				get_cat_count_sql += ` WHERE julianday('now') - julianday(submit_date) <= 31`
			case "year":
				get_cat_count_sql += ` WHERE julianday('now') - julianday(submit_date) <= 366`
		}

		var c sql.NullInt32
		err = db.QueryRow(get_cat_count_sql).Scan(&c)
		if err != nil {
			log.Fatal(err)
		}

		category_counts[i].Count = c.Int32
	}

	slices.SortFunc(category_counts, model.SortCategories)

	// return top {LIMIT} categories and their counts
	render.Status(r, http.StatusOK)
	render.JSON(w, r, category_counts[0:LIMIT])
}

// ADD NEW TAG
func AddTag(w http.ResponseWriter, r *http.Request) {
	tag_data := &model.NewTagRequest{}
	if err := render.Bind(r, tag_data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Check auth token
	var req_login_name string
	claims, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if len(claims) > 0 {
		req_login_name = claims["login_name"].(string)
	}

	// Check that tag has no more than 5 (for now) categories
	// e.g., history,science,politics,funny,internet
	cat_limit := 5
	if strings.Count(tag_data.Categories, ",") > cat_limit {
		render.Render(w, r, ErrInvalidRequest(errors.New("tag has too many categories")))
		return
	}

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check if link exists, Abort if invalid link ID provided
	var s sql.NullString
	err = db.QueryRow("SELECT id FROM Links WHERE id = ?;", tag_data.LinkID).Scan(&s)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("invalid link id provided")))
		return
	}

	// Check if duplicate (same link ID, submitted by), Abort if so
	err = db.QueryRow("SELECT id FROM Tags WHERE link_id = ? AND submitted_by = ?;", tag_data.LinkID, req_login_name).Scan(&s)
	if err == nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("duplicate tag")))
		return
	}

	// Convert tag categories to lowercase
	tag_data.Categories = strings.ToLower(tag_data.Categories)

	// Insert new tag
	res, err := db.Exec("INSERT INTO Tags VALUES(?,?,?,?,?);", nil, tag_data.LinkID, tag_data.Categories, req_login_name, tag_data.LastUpdated)
	if err != nil {
		log.Fatal(err)
	}

	// Recalculate global categories for this link
	RecalculateGlobalCategories(db, tag_data.LinkID)

	var id int64
	if id, err = res.LastInsertId(); err != nil {
		log.Fatal(err)
	}
	tag_data.ID = id

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, tag_data)
}

// EDIT TAG
func EditTag(w http.ResponseWriter, r *http.Request) {
	edit_tag_data := &model.EditTagRequest{}
	if err := render.Bind(r, edit_tag_data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Check auth token
	var req_login_name string
	claims, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if len(claims) > 0 {
		req_login_name = claims["login_name"].(string)
	}

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check if tag doesn't exist or submitted by a different user, Abort if either
	var t sql.NullString
	err = db.QueryRow("SELECT submitted_by FROM Tags WHERE id = ?;", edit_tag_data.ID).Scan(&t)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("tag not found")))
		return
	} else if t.String != req_login_name {
		render.Render(w, r, ErrInvalidRequest(errors.New("cannot edit another user's tag")))
		return
	}

	// Sort categories alphabetically
	split_categories := strings.Split(edit_tag_data.Categories, ",")
	slices.Sort(split_categories)
	edit_tag_data.Categories = strings.Join(split_categories, ",")

	_, err = db.Exec(`UPDATE Tags 
	SET categories = ?, last_updated = ? 
	WHERE id = ?;`, 
	edit_tag_data.Categories, edit_tag_data.LastUpdated, edit_tag_data.ID)
	if err != nil {
		log.Fatal(err)
	}

	// Get link ID from tag ID
	var lid sql.NullString
	err = db.QueryRow("SELECT link_id FROM Tags WHERE id = ?;", edit_tag_data.ID).Scan(&lid)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("invalid tag id provided")))
		return
	}

	// Recalculate global categories for this link
	RecalculateGlobalCategories(db, lid.String)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, edit_tag_data)

}

// Recalculate global categories for a link whose tags changed
func RecalculateGlobalCategories(db *sql.DB, link_id string) {
	// (technically should affect all links that share 1+ categories but that's too complicated.) 
	// (Plus, many links will not be seen enough to justify being updated constantly. Makes enough sense to only update a link's global cats when a new tag is added to that link.)

	category_scores := make(map[string]float32)
	// Which tags have the earliest last_updated of this link's tags? (top 20)
	// (in other words, occupying the greatest % of the link's lifespan without revision)
	// What are the categories of those tags?

	rows, err := db.Query(`SELECT (julianday('now') - julianday(last_updated)) / (julianday('now') - julianday(submit_date)) AS lifespan_overlap, categories 
		FROM Tags 
		INNER JOIN Links 
		ON Links.id = Tags.link_id 
		WHERE link_id = ? 
		ORDER BY lifespan_overlap DESC 
		LIMIT 20;`, link_id)
	if err != nil {
		log.Fatal(err)
	}

	earliest_tags := []model.EarlyTag{}
	for rows.Next() {
		var t model.EarlyTag
		err = rows.Scan(&t.LifeSpanOverlap, &t.Categories)
		if err != nil {
			log.Fatal(err)
		}
		earliest_tags = append(earliest_tags, t)
	}

	// 50% Max category score from across all tags used as threshold for assignment to Global Tag categories
	var max_cat_score float32

	// Tag score limit determined by number of tags so combined scores always sum to 1
	tag_score_limit := 1 / float32(len(earliest_tags))
	for _, tag := range earliest_tags {

		// convert to all lowercase
		lc := strings.ToLower(tag.Categories)

		// take square root of lifespan overlap to smooth out scores
		// (allow brand-new tags to still have some influence)
		// e.g. sqrt(0.01) = 0.1
		tag.LifeSpanOverlap = float32(math.Sqrt(float64(tag.LifeSpanOverlap)))

		// add scores for each category if multiple
		// Note: categories that appear multiple times across different tags will have multipled scores
		// (more likely to affect Global Tag categories)
		if strings.Contains(tag.Categories, ",") {
			c := strings.Split(lc, ",")

			for _, cat := range c {
				category_scores[cat] += tag.LifeSpanOverlap * tag_score_limit

				// update max score (to be used when assigning global categories)
				if category_scores[cat] > max_cat_score {
					max_cat_score = category_scores[cat]
				}
			}

		// else add score for single category
		} else {
			category_scores[lc] += tag.LifeSpanOverlap * tag_score_limit

			// update max score
			if category_scores[lc] > max_cat_score {
				max_cat_score = category_scores[lc]
			}
		}
	}

	// Sort categories alphabetically
	sorted_cats := make([]string, 0, len(category_scores))
	for cat := range category_scores {
		sorted_cats = append(sorted_cats, cat)
	}
	slices.Sort(sorted_cats)
	
	// Assign categories scoring 50%+ of max score to Global Tag
	var global_cats string
	for _, cat := range sorted_cats {
		if category_scores[cat] >= 0.5*max_cat_score {
			global_cats += cat + ","
		}
	}

	// Remove trailing comma
	if len(global_cats) > 0 {
		global_cats = global_cats[:len(global_cats)-1]
	}

	// Update link
	_, err = db.Exec("UPDATE Links SET global_cats = ? WHERE id = ?;", global_cats, link_id)
	if err != nil {
		log.Fatal(err)
	}
}