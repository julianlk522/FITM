package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/jwtauth/v5"
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
	_, claims, err := jwtauth.FromContext(r.Context())
	// claims = {"user_id":"1234","login_name":"johndoe"}
	if err != nil {
		log.Fatal(err)
	}
	req_login_name, ok := claims["login_name"]
	if !ok {
		log.Fatal("invalid auth token")
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
	model.RecalcGlobalCats(db, tag_data.LinkID)

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
	_, claims, err := jwtauth.FromContext(r.Context())
	// claims = {"user_id":"1234","login_name":"johndoe"}
	if err != nil {
		log.Fatal(err)
	}
	req_login_name, ok := claims["login_name"]
	if !ok {
		log.Fatal("invalid auth token")
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
	model.RecalcGlobalCats(db, lid.String)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, edit_tag_data)

}