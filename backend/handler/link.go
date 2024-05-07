package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"oitm/model"
)

// GET OVERALL MOST-LIKED LINKS
// (top 20 for now)
func GetTopLinks(w http.ResponseWriter, r *http.Request) {
	db ,err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	get_link_likes_sql := `SELECT Links.id as link_id, url, submitted_by, submit_date, coalesce(like_count,0) as like_count FROM LINKS LEFT JOIN (SELECT link_id as likes_link_id, count(*) as like_count FROM 'Link Likes' GROUP BY likes_link_id) ON Links.id = likes_link_id ORDER BY like_count DESC, link_id ASC;`

	links := []model.Link{}
	rows, err := db.Query(get_link_likes_sql)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		i := model.Link{}
		err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.LikeCount)
		if err != nil {
			log.Fatal(err)
		}
		links = append(links, i)
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, links)
}

// GET MOST-LIKED LINKS DURING PERIOD
// (day, week, month)
// (top 20 for now)
func GetTopLinksByPeriod(w http.ResponseWriter, r *http.Request) {
	db ,err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	get_link_likes_sql := `SELECT Links.id as link_id, url, submitted_by, submit_date, coalesce(like_count,0) as like_count FROM LINKS LEFT JOIN (SELECT link_id as likes_link_id, count(*) as like_count FROM 'Link Likes' GROUP BY likes_link_id) ON Links.id = likes_link_id`

	switch chi.URLParam(r, "period") {
	case "day":
		get_link_likes_sql += ` WHERE julianday('now') - julianday(submit_date) <= 2`
	case "week":
		get_link_likes_sql += ` WHERE julianday('now') - julianday(submit_date) <= 8`
	case "month":
		get_link_likes_sql += ` WHERE julianday('now') - julianday(submit_date) <= 31`
	default:
		render.Render(w, r, ErrInvalidRequest(errors.New("invalid period")))
		return
	}

	get_link_likes_sql += ` ORDER BY like_count DESC, link_id ASC;`

	links := []model.Link{}
	rows, err := db.Query(get_link_likes_sql)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		i := model.Link{}
		err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.LikeCount)
		if err != nil {
			log.Fatal(err)
		}
		links = append(links, i)
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, links)
}

// GET MOST-LIKED LINKS WITH GIVEN CATEGORY(IES)
// (top 20 for now)
func GetTopLinksByCategories(w http.ResponseWriter, r *http.Request) {
	db ,err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// get categories
	categories_params := chi.URLParam(r, "categories")
	var get_links_sql string

	// multiple categories
	if strings.Contains(categories_params, ",") {
		categories := strings.Split(categories_params, ",")

		// get link IDs
		get_links_sql = fmt.Sprintf(`select id from Links where ',' || global_cats || ',' like '%%,%s,%%'`, categories[0])

		for i := 1; i < len(categories); i++ {
			get_links_sql += fmt.Sprintf(` AND ',' || global_cats || ',' like '%%,%s,%%'`, categories[i])
		}

		get_links_sql += ` group by link_id`
	// single category
	} else {

		// get link IDs
		get_links_sql = fmt.Sprintf(`select id from Links where ',' || global_cats || ',' like '%%,%s,%%' group by id`, categories_params)
	}

	rows, err := db.Query(get_links_sql)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	
	var link_ids []string
	for rows.Next() {
		var link_id string
		err := rows.Scan(&link_id)
		if err != nil {
			log.Fatal(err)
		}
		link_ids = append(link_ids, link_id)
	}

	// if no links found, return status message
	if len(link_ids) == 0 {
		return_json := map[string]string{
			"message": "no links found",
		}
		render.JSON(w, r, return_json)
		render.Status(r, http.StatusNoContent)
		return
	}

	// get total likes for each link_id
	db, err = sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err = db.Query(fmt.Sprintf(`SELECT Links.id as link_id, url, submitted_by, submit_date, coalesce(global_cats,"") as categories, coalesce(like_count,0) as like_count FROM LINKS LEFT JOIN (SELECT link_id as likes_link_id, count(*) as like_count FROM 'Link Likes' GROUP BY likes_link_id) ON Links.id = likes_link_id WHERE link_id IN (%s) ORDER BY like_count DESC, link_id ASC;`, strings.Join(link_ids, ",")))
	if err != nil {
		log.Fatal(err)
	}

	links := []model.Link{}
	for rows.Next() {
		i := model.Link{}
		err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.Categories, &i.LikeCount)
		if err != nil {
			log.Fatal(err)
		}
		links = append(links, i)
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, links)

}

// ADD NEW LINK
func AddLink(w http.ResponseWriter, r *http.Request) {
	link_data := &model.NewLinkRequest{}
	if err := render.Bind(r, link_data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// TODO: check auth token

	// Check if link exists, Abort if attempting duplicate
	var s sql.NullString
	err = db.QueryRow("SELECT url FROM Links WHERE url = ?", link_data.URL).Scan(&s)
	if err == nil {
		// note: use this error
		render.Render(w, r, ErrInvalidRequest(errors.New("link already exists")))
		return
	}

	res, err := db.Exec("INSERT INTO Links VALUES(?,?,?,?,?);", nil, link_data.URL, link_data.SubmittedBy, link_data.SubmitDate, "")
	if err != nil {
		log.Fatal(err)
	}

	var id int64
	if id, err = res.LastInsertId(); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
	}
	link_data.ID = id

	// Insert new tag
	_, err = db.Exec("INSERT INTO Tags VALUES(?,?,?,?,?);", nil, link_data.ID, link_data.Categories, link_data.SubmittedBy, link_data.SubmitDate)
	if err != nil {
		log.Fatal(err)
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, link_data)
}

// COPY LINK TO USER'S TREASURE MAP
func CopyLinkToMap(w http.ResponseWriter, r *http.Request) {
	copy_link_data := &model.LinkCopyRequest{}
	if err := render.Bind(r, copy_link_data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	res, err := db.Exec("INSERT INTO 'Link Copies' VALUES(?,?,?);", nil, copy_link_data.LinkID, copy_link_data.UserID)
	if err != nil {
		log.Fatal(err)
	}

	var id int64
	if id, err = res.LastInsertId(); err != nil {
		log.Fatal(err)
	}

	return_json := map[string]int64{
		"copy_id": id,
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, return_json)
}

// UN-COPY LINK
func UncopyLink(w http.ResponseWriter, r *http.Request) {
	delete_copy_data := &model.DeleteLinkCopyRequest{}
	if err := render.Bind(r, delete_copy_data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check if link copy exists
	var s sql.NullString
	err = db.QueryRow("SELECT id FROM 'Link Copies' WHERE id = ?;", delete_copy_data.ID).Scan(&s)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("link copy does not exist")))
		return
	}

	// Todo: check auth token, ensure that user is owner of link copy

	// Delete
	_, err = db.Exec("DELETE FROM 'Link Copies' WHERE id = ?;", delete_copy_data.ID)
	if err != nil {
		log.Fatal(err)
	}

	return_json := map[string]string{
		"status": "success",
	}

	render.JSON(w, r, return_json)
	render.Status(r, http.StatusNoContent)
}

// GET LINK LIKES
// (not currently used - probably delete later)
func GetLinkLikes(w http.ResponseWriter, r *http.Request) {
	link_id := chi.URLParam(r, "id")
	if link_id == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("invalid link id provided")))
		return
	}

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	// Check if link exists, Abort if invalid link ID provided
	var s sql.NullString
	err = db.QueryRow("SELECT id FROM Links WHERE id = ?;", link_id).Scan(&s)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("no link found with given ID")))
		return
	}

	// Get like count
	var c int64
	err = db.QueryRow("SELECT COUNT(id) as count FROM 'Link Likes' WHERE link_id = ?;", link_id).Scan(&c)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	return_json := map[string]int64{"likes": c}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, return_json)
}