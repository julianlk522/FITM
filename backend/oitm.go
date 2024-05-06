package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/slices"
)

func main() {	
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Home - check if server running
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World!")
	})

	// USER ACCOUNTS
	// Sign Up
	r.Post("/signup", func(w http.ResponseWriter, r *http.Request) {
		signup_data := &SignUpRequest{}

		if err := render.Bind(r, signup_data); err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}

		db, err := sql.Open("sqlite3", "./db/oitm.db")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		// Check if user already exists, Abort if so
		var s sql.NullString
		err = db.QueryRow("SELECT login_name FROM Users WHERE login_name = ?", signup_data.UserAuth.LoginName).Scan(&s)
		if err == nil {
			render.Render(w, r, ErrInvalidRequest(errors.New("login name taken")))
			return
		}

		// Hash password
		pw_hash, err := bcrypt.GenerateFromPassword([]byte(signup_data.UserAuth.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal(err)
		}

		_, err = db.Exec(`INSERT INTO users VALUES (?,?,?,?,?,?)`, nil, signup_data.UserAuth.LoginName, pw_hash, nil, nil, signup_data.Created)
		if err != nil {
			log.Fatal(err)
		}

		// TODO: generate and return jwt along with login name
		var token string = "token"
		return_json := map[string]string{"token": token, "login_name": signup_data.UserAuth.LoginName}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, return_json)
	})

	// Log In
	r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		login_data := &LogInRequest{}

		if err := render.Bind(r, login_data); err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}

		db, err := sql.Open("sqlite3", "./db/oitm.db")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		// Attempt to collect user hashed password, 
		// Abort if user not found
		var p sql.NullString
		err = db.QueryRow("SELECT password FROM Users WHERE login_name = ?", login_data.LoginName).Scan(&p)
		if err != nil {
			render.Render(w, r, ErrInvalidRequest(errors.New("no user found with given login name")))
			return
		}

		// compare password hashes
		err = bcrypt.CompareHashAndPassword([]byte(p.String), []byte(login_data.UserAuth.Password))
		if err != nil {
			render.Render(w, r, ErrInvalidRequest(errors.New("incorrect password")))
			return
		}

		// TODO: generate and return jwt along with login name
		var token string = "token"
		return_json := map[string]string{"token": token, "login_name": login_data.LoginName}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, return_json)
	})

	// Edit profile (about or profile pic)
	r.Patch("/users", func(w http.ResponseWriter, r *http.Request) {
		edit_profile_data := &EditProfileRequest{}

		if err := render.Bind(r, edit_profile_data); err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}

		db, err := sql.Open("sqlite3", "./db/oitm.db")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		
		return_json := map[string]string{"token": edit_profile_data.AuthToken}
		
		// TODO: check auth token

		// About
		if edit_profile_data.EditAboutRequest != nil {
			// TODO replace hard-coded id with id corresponding
			// to provided auth token
			_, err = db.Exec(`UPDATE Users SET about = ? WHERE id = ?`, edit_profile_data.EditAboutRequest.About, edit_profile_data.AuthToken)
			if err != nil {
				log.Fatal(err)
			}

			return_json["about"] = edit_profile_data.EditAboutRequest.About
		}
		
		// Profile Pic
		if edit_profile_data.EditPfpRequest != nil {
			// TODO replace hard-coded id with id corresponding
			// to provided auth token
			_, err = db.Exec(`UPDATE Users SET pfp = ? WHERE id = ?`, edit_profile_data.EditPfpRequest.PFP, edit_profile_data.AuthToken)
			if err != nil {
				log.Fatal(err)
			}

			return_json["pfp"] = edit_profile_data.EditPfpRequest.PFP
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, return_json)
	})

	// Get user treasure map
	// (includes links tagged by and copied by user)
	r.Get("/map/{id}", func(w http.ResponseWriter, r *http.Request) {
		var user_id string = chi.URLParam(r, "id")

		db ,err := sql.Open("sqlite3", "./db/oitm.db")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		// get user-assigned tag categories
		get_custom_cats_sql := fmt.Sprintf(`SELECT link_id, categories as cats FROM Tags JOIN Users
		ON Users.login_name = Tags.submitted_by
		WHERE Users.id = '%s'`, user_id)

		rows, err := db.Query(get_custom_cats_sql)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		user_custom_cats := []CustomLinkCategories{}
		for rows.Next() {
			i := CustomLinkCategories{}
			err := rows.Scan(&i.LinkID, &i.Categories)
			if err != nil {
				log.Fatal(err)
			}
			user_custom_cats = append(user_custom_cats, i)
		}

		// get all map links and their global categories + like counts
		get_map_sql := fmt.Sprintf(`SELECT Links.id as link_id, url, submitted_by, submit_date, coalesce(global_cats,"") as global_cats, coalesce(like_count,0) as like_count FROM LINKS LEFT JOIN (SELECT link_id as like_link_id, count(*) as like_count FROM 'Link Likes' GROUP BY like_link_id) ON Links.id = like_link_id WHERE link_id IN ( SELECT link_id FROM Tags JOIN Users ON Users.login_name = Tags.submitted_by WHERE Users.id = '%s' UNION SELECT link_id FROM (SELECT link_id, NULL as cats, user_id as link_copier_id FROM 'Link Copies' JOIN Users ON Users.id = link_copier_id WHERE link_copier_id = '%s') JOIN Links ON Links.id = link_id) ORDER BY like_count DESC, link_id ASC;`, user_id, user_id)

		links := []Link{}
		rows, err = db.Query(get_map_sql)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		for rows.Next() {
			i := Link{}
			err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.Categories, &i.LikeCount)
			if err != nil {
				log.Fatal(err)
			}
			links = append(links, i)
		}

		// replace global categories for links to which the user has submitted their own tags
		for i, link := range links {
			for _, cat := range user_custom_cats {
				if link.ID == cat.LinkID {
					links[i].Categories = cat.Categories
				}
			}
		}

		render.JSON(w, r, links)
		render.Status(r, http.StatusOK)
	})

	// LINKS
	// Get most-liked links overall
	// (top 20 for now)
	r.Get("/links", func(w http.ResponseWriter, r *http.Request) {
		db ,err := sql.Open("sqlite3", "./db/oitm.db")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		get_link_likes_sql := `SELECT Links.id as link_id, url, submitted_by, submit_date, coalesce(like_count,0) as like_count FROM LINKS LEFT JOIN (SELECT link_id as likes_link_id, count(*) as like_count FROM 'Link Likes' GROUP BY likes_link_id) ON Links.id = likes_link_id ORDER BY like_count DESC, link_id ASC;`

		links := []Link{}
		rows, err := db.Query(get_link_likes_sql)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		for rows.Next() {
			i := Link{}
			err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.LikeCount)
			if err != nil {
				log.Fatal(err)
			}
			links = append(links, i)
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, links)
	})

	// Get most-liked links during given period
	// (day, week, month)
	// (top 20 for now)
	r.Get("/links/{period}", func(w http.ResponseWriter, r *http.Request) {
		db ,err := sql.Open("sqlite3", "./db/oitm.db")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		get_link_likes_sql := `SELECT Links.id as link_id, url, submitted_by, submit_date, coalesce(like_count,0) as like_count FROM LINKS LEFT JOIN (SELECT link_id as likes_link_id, count(*) as like_count FROM 'Link Likes' GROUP BY likes_link_id) ON Links.id = likes_link_id ORDER BY like_count DESC, link_id ASC`

		switch chi.URLParam(r, "period") {
		case "day":
			get_link_likes_sql += ` WHERE julianday('now') - julianday(submit_date) <= 2;`
		case "week":
			get_link_likes_sql += ` WHERE julianday('now') - julianday(submit_date) <= 8;`
		case "month":
			get_link_likes_sql += ` WHERE julianday('now') - julianday(submit_date) <= 31;`
		default:
			render.Render(w, r, ErrInvalidRequest(errors.New("invalid period")))
			return
		}

		links := []Link{}
		rows, err := db.Query(get_link_likes_sql)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		for rows.Next() {
			i := Link{}
			err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.LikeCount)
			if err != nil {
				log.Fatal(err)
			}
			links = append(links, i)
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, links)
	})

	// Get most-liked links with given category(ies) on the global map
	// (top 20 for now)
	r.Get("/links/cat/{categories}", func(w http.ResponseWriter, r *http.Request) {
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

		links := []Link{}
		for rows.Next() {
			i := Link{}
			err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.Categories, &i.LikeCount)
			if err != nil {
				log.Fatal(err)
			}
			links = append(links, i)
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, links)

	})

	// Add New Link
	r.Post("/links", func(w http.ResponseWriter, r *http.Request) {
		link_data := &NewLinkRequest{}
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
			render.Render(w, r, ErrInvalidRequest(errors.New("Link already exists")))
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

	})

	// Copy link to user's treasure map
	r.Post("/links/copy", func(w http.ResponseWriter, r *http.Request) {
		copy_link_data := &LinkCopyRequest{}
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
	})

	// Remove copied link from user's treasure map
	r.Delete("/links/copy", func(w http.ResponseWriter, r *http.Request) {
		delete_copy_data := &DeleteLinkCopyRequest{}
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
			render.Render(w, r, ErrInvalidRequest(errors.New("Link copy does not exist")))
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
	})

	// Get link likes
	// (not currently used, likely delete but keep for reference)
	r.Get("/links/{id}/likes", func(w http.ResponseWriter, r *http.Request) {
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
	})

	// TAGS
	// Get Most-Used Tag Categories
	// Todo: edit to search global categories instead
	r.Get("/tags/popular", func(w http.ResponseWriter, r *http.Request) {

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
		var category_counts []CategoryCount = make([]CategoryCount, len(categories))
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

		slices.SortFunc(category_counts, sort_categories)

		// return top {LIMIT} categories and their counts
		render.Status(r, http.StatusOK)
		render.JSON(w, r, category_counts[0:LIMIT])
	})

	// Add New Tag
	r.Post("/tags", func(w http.ResponseWriter, r *http.Request) {
		tag_data := &NewTagRequest{}
		if err := render.Bind(r, tag_data); err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
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
		err = db.QueryRow("SELECT id FROM Tags WHERE link_id = ? AND submitted_by = ?;", tag_data.LinkID, tag_data.SubmittedBy).Scan(&s)
		if err == nil {
			render.Render(w, r, ErrInvalidRequest(errors.New("duplicate tag")))
			return
		}

		// Convert tag categories to lowercase
		tag_data.Categories = strings.ToLower(tag_data.Categories)

		// Insert new tag
		res, err := db.Exec("INSERT INTO Tags VALUES(?,?,?,?,?);", nil, tag_data.LinkID, tag_data.Categories, tag_data.SubmittedBy, tag_data.LastUpdated)
		if err != nil {
			log.Fatal(err)
		}

		// Recalculate global categories for this link
		recalc_global_cats(db, tag_data.LinkID)

		var id int64
		if id, err = res.LastInsertId(); err != nil {
			log.Fatal(err)
		}
		tag_data.ID = id

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, tag_data)
	})

	// Edit Tag
	r.Put("/tags", func(w http.ResponseWriter, r *http.Request) {
		edit_tag_data := &EditTagRequest{}
		if err := render.Bind(r, edit_tag_data); err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}

		db, err := sql.Open("sqlite3", "./db/oitm.db")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		// TODO: check auth token

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
		recalc_global_cats(db, lid.String)

		render.Status(r, http.StatusOK)
		render.JSON(w, r, edit_tag_data)

	})

	// SUMMARIES
	// Create Summary / Like Summary
	r.Post("/summaries", func(w http.ResponseWriter, r *http.Request) {
		summary_data := &SummaryRequest{}

		if err := render.Bind(r, summary_data); err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}

		db, err := sql.Open("sqlite3", "./db/oitm.db")
		if err != nil {
			log.Fatal(err)
		}

		defer db.Close()

		// Create Summary
		if summary_data.NewSummaryRequest != nil {

			// Check if link exists, Abort if not
			var s sql.NullString
			err = db.QueryRow("SELECT id FROM Links WHERE id = ?", summary_data.LinkID).Scan(&s)
			if err != nil {
				render.Render(w, r, ErrInvalidRequest(errors.New("link not found")))
				return
			}

			// TODO: check auth token

			_, err = db.Exec(`INSERT INTO Summaries VALUES (?,?,?,?)`, nil, summary_data.NewSummaryRequest.Text, summary_data.NewSummaryRequest.LinkID, summary_data.NewSummaryRequest.SubmittedBy)
			if err != nil {
				log.Fatal(err)
			}

		// Like Summary
		} else if summary_data.NewSummaryLikeRequest != nil {

			// Check if summary exists, Abort if not
			var s sql.NullString
			err = db.QueryRow("SELECT id FROM Summaries WHERE id = ?", summary_data.NewSummaryLikeRequest.SummaryID).Scan(&s)
			if err != nil {
				render.Render(w, r, ErrInvalidRequest(errors.New("summary not found")))
				return
			}

			// TODO: check auth token

			_, err = db.Exec(`INSERT INTO 'Summary Likes' VALUES (?,?,?)`, nil, summary_data.NewSummaryLikeRequest.UserID, summary_data.NewSummaryLikeRequest.SummaryID)
			if err != nil {
				log.Fatal(err)
			}
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, summary_data)
	})

	// Edit Summary
	r.Patch("/summaries", func(w http.ResponseWriter, r *http.Request) {
		edit_data := &SummaryRequest{}

		if err := render.Bind(r, edit_data); err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}

		db, err := sql.Open("sqlite3", "./db/oitm.db")
		if err != nil {
			log.Fatal(err)
		}

		defer db.Close()

		// TODO: check auth token

		// Check if summary exists, Abort if not
		var s sql.NullString
		err = db.QueryRow("SELECT id FROM Summaries WHERE id = ?", edit_data.EditSummaryRequest.SummaryID).Scan(&s)
		if err != nil {
			render.Render(w, r, ErrInvalidRequest(errors.New("summary not found")))
			return
		}

		_, err = db.Exec(`UPDATE Summaries SET text = ? WHERE id = ?`, edit_data.EditSummaryRequest.Text, edit_data.EditSummaryRequest.SummaryID)
		if err != nil {
			log.Fatal(err)
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, edit_data)
	})

	// Delete Summary / Unlike Summary
	r.Delete("/summaries", func(w http.ResponseWriter, r *http.Request) {
		summary_data := &SummaryRequest{}

		if err := render.Bind(r, summary_data); err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}

		db, err := sql.Open("sqlite3", "./db/oitm.db")
		if err != nil {
			log.Fatal(err)
		}

		defer db.Close()

		// Delete Summary
		if summary_data.DeleteSummaryRequest != nil {
			
			// TODO: check auth token

			_, err = db.Exec(`DELETE FROM Summaries WHERE id = ?`, summary_data.DeleteSummaryRequest.SummaryID)
			if err != nil {
				log.Fatal(err)
			}

		// Unlike Summary
		} else if summary_data.DeleteSummaryLikeRequest != nil {
			
			// TODO: check auth token

			_, err = db.Exec(`DELETE FROM 'Summary Likes' WHERE id = ?`, summary_data.DeleteSummaryLikeRequest.SummaryLikeID)
			if err != nil {
				log.Fatal(err)
			}
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, map[string]string{"status": "accepted"})

	})

	// Serve
	// make sure this runs AFTER all routes
	if err := http.ListenAndServe("localhost:8000", r); err != nil {
		log.Fatal(err)
	}
}

// Recalculate global categories for a link whose tags changed
 func recalc_global_cats(db *sql.DB, link_id string) {
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

	earliest_tags := []EarliestTagCats{}
	for rows.Next() {
		var t EarliestTagCats
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

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

var ErrNotFound = &ErrResponse{HTTPStatusCode: 404, StatusText: "Resource not found."}

// TYPES

// USER
type UserAuth struct {
	LoginName string `json:"login_name"`
	Password string `json:"password"`
}

type User struct {
	*UserAuth
	ID int64
	About string
	ProfilePic string
}
type SignUpRequest struct {
	*UserAuth
	Created string
}

func (a *SignUpRequest) Bind(r *http.Request) error {
	if a.UserAuth == nil {
		return errors.New("signup info not provided")
	} else if a.UserAuth.LoginName == "" {
		return errors.New("missing login name")
	} else if a.UserAuth.Password == "" {
		return errors.New("missing password")
	}

	a.Created = time.Now().Format("2006-01-02 15:04:05")
	return nil
}

type LogInRequest struct {
	*UserAuth
}


func (a *LogInRequest) Bind(r *http.Request) error {
	if a.UserAuth == nil {
		return errors.New("login info not provided")
	} else if a.UserAuth.LoginName == "" {
		return errors.New("missing login name")
	} else if a.UserAuth.Password == "" {
		return errors.New("missing password")
	}

	return nil
}

type EditProfileRequest struct {
	AuthToken string `json:"token"`
	*EditAboutRequest
	*EditPfpRequest
}

func (a *EditProfileRequest) Bind(r *http.Request) error {
	if a.AuthToken == "" {
		return errors.New("missing auth token")
	}

	// TODO: check auth token

	return nil
}

type EditAboutRequest struct {
	About string `json:"about,omitempty"`
}

type EditPfpRequest struct {
	PFP string `json:"pfp,omitempty"`
}

type CustomLinkCategories struct {
	LinkID int64
	Categories string
}

// LINK
type Link struct {
	ID int64
	URL string
	SubmittedBy string
	SubmitDate string
	Categories string
	LikeCount int64
}

type NewLink struct {
	URL string `json:"url"`
	SubmittedBy string `json:"submitted_by"`
	Categories string `json:"categories"`
}

type NewLinkRequest struct {
	*NewLink
	ID int64
	SubmitDate string
}

func (a *NewLinkRequest) Bind(r *http.Request) error {
	if a.NewLink == nil {
		return errors.New("missing required Link fields")
	}

	a.SubmitDate = time.Now().Format("2006-01-02 15:04:05")

	return nil
}

type LinkCopyRequest struct {
	ID int64
	AuthToken string `json:"token"`
	LinkID string `json:"link_id"`
	UserID string `json:"user_id"`
}

func (a *LinkCopyRequest) Bind(r *http.Request) error {
	if a.AuthToken == "" {
		return errors.New("missing auth token")
	}
	if a.LinkID == "" {
		return errors.New("missing link ID")
	}
	if a.UserID == "" {
		return errors.New("missing user ID")
	}

	return nil
}

type DeleteLinkCopyRequest struct {
	ID string `json:"copy_id"`
	AuthToken string `json:"token"`
}

func (a *DeleteLinkCopyRequest) Bind(r *http.Request) error {
	if a.ID == "" {
		return errors.New("missing copy ID")
	}
	if a.AuthToken == "" {
		return errors.New("missing auth token")
	}

	return nil
}

// TAG
type NewTag struct {
	LinkID string `json:"link_id"`
	Categories string `json:"categories"`
	SubmittedBy string `json:"submitted_by"`
}

type NewTagRequest struct {
	*NewTag
	ID int64
	LastUpdated string
}

func (a *NewTagRequest) Bind(r *http.Request) error {
	if a.NewTag == nil {
		return errors.New("missing required Tag fields")
	}

	a.LastUpdated = time.Now().Format("2006-01-02 15:04:05")

	return nil
}

type EditTagRequest struct {
	AuthToken string `json:"token"`
	ID string `json:"tag_id"`
	Categories string `json:"categories"`
}

func (a *EditTagRequest) Bind(r *http.Request) error {
	if a.AuthToken == "" {
		return errors.New("missing auth token")
	}
	if a.ID == "" {
		return errors.New("missing tag ID")
	}
	if a.Categories == "" {
		return errors.New("missing categories")
	}

	return nil
}

type CategoryCount struct {
	Category string
	Count int32
}

func sort_categories(i, j CategoryCount) int {
	if i.Count > j.Count {
		return -1
	} else if i.Count == j.Count && i.Category < j.Category {
		return -1
	}
	return 1
}

type EarliestTagCats struct {
	LifeSpanOverlap float32
	Categories string
}

// SUMMARY
type SummaryRequest struct {
	*NewSummaryRequest
	*EditSummaryRequest
	*DeleteSummaryRequest
	*NewSummaryLikeRequest
	*DeleteSummaryLikeRequest
}

func (a *SummaryRequest) Bind(r *http.Request) error {
	if a.NewSummaryRequest == nil && a.NewSummaryLikeRequest == nil && a.EditSummaryRequest == nil && a.DeleteSummaryRequest == nil && a.DeleteSummaryLikeRequest == nil {
		return errors.New("missing required Summary fields")
	}

	if a.EditSummaryRequest != nil {
		if a.EditSummaryRequest.Text == "" {
			return errors.New("missing replacement summary text")
		} else if a.EditSummaryRequest.SummaryID == "" {
			return errors.New("missing summary ID")
		}
	}

	return nil
}

type NewSummaryRequest struct {
	SubmittedBy string `json:"submitted_by"`
	LinkID string `json:"link_id"`
	Text string `json:"text"`
}

type EditSummaryRequest struct {
	// would use json:"summary_id" here but conflicts with
	// below NewSummaryLikeRequest json ... not sure how else to fix
	SummaryID string `json:"summary_id_edit"`
	Text string `json:"text_edit"`
}

type DeleteSummaryRequest struct {
	// would use json:"summary_id" here but conflicts with
	// below NewSummaryLikeRequest json ... not sure how else to fix
	SummaryID string `json:"summary_id_del"`
}

type NewSummaryLikeRequest struct {
	SummaryID string `json:"summary_id"`
	UserID string `json:"user_id"`
}

type DeleteSummaryLikeRequest struct {
	SummaryLikeID string `json:"slike_id"`
}

// ERROR RESPONSE
type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}