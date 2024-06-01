package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/render"
	"golang.org/x/exp/slices"

	"oitm/model"
)

// GET MOST-USED TAG CATEGORIES
// Todo: edit to search global categories instead
func GetTopTagCategories(w http.ResponseWriter, r *http.Request) {

	// Limit 5 for now
	const LIMIT int = 5

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// get all categories
	rows, err := db.Query("select categories from tags GROUP BY categories;")
	if err != nil {
		log.Fatal(err)
	}

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

		get_cat_count_sql := fmt.Sprintf(`select count(*) as count_with_cat from (select link_id from Tags where ',' || categories || ',' like '%%,%s,%%' group by Tags.id)`, categories[i])

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

	_, err = db.Exec("UPDATE Tags SET categories = ?, last_updated = ? WHERE id = ?;", edit_tag_data.Categories, time.Now().Format("2006-01-02 15:04:05"), edit_tag_data.ID)
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

	// Global category(ies) based on aggregated scores from all tags of the link, based on time between link creation and tag creation/last edit
	category_scores := make(map[string]float32)

	// which tags have the earliest last_updated of this link's tags?
	// (in other words, occupying the greatest % of the link's lifetime without needing revision)
	// what are the categories of those tags? (top 20)
	rows, err := db.Query(`select (julianday('now') - julianday(last_updated)) / (julianday('now') - julianday(submit_date)) as prcnt_lo, categories from Tags INNER JOIN Links on Links.id = Tags.link_id WHERE link_id = ? ORDER BY prcnt_lo DESC LIMIT 20;`, link_id)
	if err != nil {
		log.Fatal(err)
	}

	earliest_tags := []model.EarliestTagCats{}
	for rows.Next() {
		var t model.EarliestTagCats
		err = rows.Scan(&t.LifeSpanOverlap, &t.Categories)
		if err != nil {
			log.Fatal(err)
		}
		earliest_tags = append(earliest_tags, t)
	}

	// add to category_scores
	var max_cat_score float32 = 0.0
	row_score_limit := 1 / float32(len(earliest_tags))
	for _, t := range earliest_tags {

		// convert to all lowercase
		lc := strings.ToLower(t.Categories)

		// use square root of life span overlap in order to smooth out scores and allow brand-new tags to still have some influence
		// e.g. sqrt(0.01) = 0.1
		t.LifeSpanOverlap = float32(math.Sqrt(float64(t.LifeSpanOverlap)))

		// split row effect among categories, if multiple
		if strings.Contains(t.Categories, ",") {
			c := strings.Split(lc, ",")
			split := float32(len(c))
			for _, cat := range c {
				category_scores[cat] += t.LifeSpanOverlap * row_score_limit / split

				// update max score (to be used when assigning global categories)
				if category_scores[cat] > max_cat_score {
					max_cat_score = category_scores[cat]
				}
			}
		} else {
			category_scores[lc] += t.LifeSpanOverlap * row_score_limit

			// update max score
			if category_scores[lc] > max_cat_score {
				max_cat_score = category_scores[lc]
			}
		}
	}

	// Determine categories with scores >= 50% of max
	var global_cats string
	for cat, score := range category_scores {
		if score >= 0.5*max_cat_score {
			global_cats += cat + ","
		}
	}
	global_cats = global_cats[:len(global_cats)-1]

	// Assign to link
	_, err = db.Exec("UPDATE Links SET global_cats = ? WHERE id = ?;", global_cats, link_id)
	if err != nil {
		log.Fatal(err)
	}
}