package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	/* Todo API Actions
	
	LINKS:
	-Copy extisting link to user's treasure map
	-Remove link from user's treasure map

	LIKES:
	-Add new like
	-Remove like

	TAGS:
	-Edit link tags
	-Add new tag category (done automatically when editing a link's tag to include a new category)

	TREASURE MAPS:
		-Get user's own treasure map
		-Get global treasure map chunks
			-intersectional reports (popular, new, etc.)
			-sectional top rankings based on likes
	
	*/
	
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Home - check if server running
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World!")
	})

	// USER ACCOUNTS
	// Sign Up
	r.Post("/users", func(w http.ResponseWriter, r *http.Request) {
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
		err = db.QueryRow("SELECT login_name FROM Users WHERE login_name = ?", signup_data.LoginName).Scan(&s)
		if err == nil {
			render.Render(w, r, ErrInvalidRequest(errors.New("login name taken")))
			return
		}

		_, err = db.Exec(`INSERT INTO users VALUES (?,?,?,?,?,?)`, nil, signup_data.LoginName, signup_data.Password, nil, nil, signup_data.Created)

		if err != nil {
			log.Fatal(err)
		}

		// TODO: generate and return jwt along with login name
		var token string = "token"
		return_json := map[string]string{"token": token, "login_name": signup_data.LoginName}

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

		// Check if user exists, Abort if not
		var s sql.NullString
		err = db.QueryRow("SELECT login_name FROM Users WHERE login_name = ?", login_data.LoginName).Scan(&s)
		if err != nil {
			render.Render(w, r, ErrInvalidRequest(errors.New("login name taken")))
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
		if summary_data.SummaryCreateRequest != nil {

			// Check if link exists, Abort if not
			var s sql.NullString
			err = db.QueryRow("SELECT id FROM Links WHERE id = ?", summary_data.LinkID).Scan(&s)
			if err != nil {
				render.Render(w, r, ErrInvalidRequest(errors.New("link not found")))
				return
			}

			// TODO: check auth token

			_, err = db.Exec(`INSERT INTO Summaries VALUES (?,?,?,?)`, nil, summary_data.SummaryCreateRequest.Text, summary_data.SummaryCreateRequest.LinkID, summary_data.SummaryCreateRequest.SubmittedBy)
			if err != nil {
				log.Fatal(err)
			}

		// Like Summary
		} else if summary_data.SummaryLikeRequest != nil {

			// Check if summary exists, Abort if not
			var s sql.NullString
			err = db.QueryRow("SELECT id FROM Summaries WHERE id = ?", summary_data.SummaryLikeRequest.SummaryID).Scan(&s)
			if err != nil {
				render.Render(w, r, ErrInvalidRequest(errors.New("summary not found")))
				return
			}

			// TODO: check auth token

			_, err = db.Exec(`INSERT INTO 'Summary Likes' VALUES (?,?,?)`, nil, summary_data.SummaryLikeRequest.UserID, summary_data.SummaryLikeRequest.SummaryID)
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
		err = db.QueryRow("SELECT id FROM Summaries WHERE id = ?", edit_data.SummaryEditRequest.SummaryID).Scan(&s)
		if err != nil {
			render.Render(w, r, ErrInvalidRequest(errors.New("summary not found")))
			return
		}

		_, err = db.Exec(`UPDATE Summaries SET text = ? WHERE id = ?`, edit_data.SummaryEditRequest.Text, edit_data.SummaryEditRequest.SummaryID)
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
		if summary_data.SummaryDeleteRequest != nil {
			
			// TODO: check auth token

			_, err = db.Exec(`DELETE FROM Summaries WHERE id = ?`, summary_data.SummaryDeleteRequest.SummaryID)
			if err != nil {
				log.Fatal(err)
			}

		// Unlike Summary
		} else if summary_data.SummaryLikeDeleteRequest != nil {
			
			// TODO: check auth token

			_, err = db.Exec(`DELETE FROM 'Summary Likes' WHERE id = ?`, summary_data.SummaryLikeDeleteRequest.SummaryLikeID)
			if err != nil {
				log.Fatal(err)
			}
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, map[string]string{"status": "accepted"})

	})

	// LINKS
	// Get All Links
	r.Get("/links", func(w http.ResponseWriter, r *http.Request) {
		db, err := sql.Open("sqlite3", "./db/oitm.db")

		if err != nil {
			log.Fatal(err)
		}

		defer db.Close()

		get_links_sql := `SELECT * FROM links`
		rows, err := db.Query(get_links_sql)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		
		links := []Link{}
		for rows.Next() {
			i := Link{}
			err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate)
			if err != nil {
				log.Fatal(err)
			}
			links = append(links, i)
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, links)
	})

	// Get most-liked links with 1+ categories on the global map
	// (top 20 for now)
	// using categories in URL parmams
	// r.Get("/links/{categories}", func(w http.ResponseWriter, r *http.Request) {

	// 	db ,err := sql.Open("sqlite3", "./db/oitm.db")
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	defer db.Close()

	// 	// get categories
	// 	categories_params := chi.URLParam(r, "categories")
	// 	var categories []string
	// 	// multiple categories
	// 	if strings.Contains(categories_params, ",") {
	// 		categories = strings.Split(categories_params, ",")
	// 		fmt.Println(categories)

	// 		get_links_sql := `SELECT * FROM Links WHERE ___`
	// 	} else {
	// 		categories = append(categories, categories_params)
	// 	}

	// })

	// Add New Link
	r.Post("/links", func(w http.ResponseWriter, r *http.Request) {
		link_data := &LinkRequest{}
		if err := render.Bind(r, link_data); err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}

		db, err := sql.Open("sqlite3", "./db/oitm.db")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		// Check if link exists, Abort if attempting duplicate
		var s sql.NullString
		err = db.QueryRow("SELECT url FROM Links WHERE url = ?", link_data.URL).Scan(&s)
		if err == nil {
			// note: use this error
			render.Render(w, r, ErrInvalidRequest(errors.New("Link already exists")))
			return
		}

		res, err := db.Exec("INSERT INTO Links VALUES(?,?,?,?);", nil, link_data.URL, link_data.SubmittedBy, link_data.SubmitDate)
		if err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
		}

		var id int64
		if id, err = res.LastInsertId(); err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
		}
		link_data.ID = id

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, link_data)

	})

	// Get link likes
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
	// Add New Tag
	r.Post("/tags", func(w http.ResponseWriter, r *http.Request) {
		tag_data := &TagCreateRequest{}
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

		// Insert new tag
		// Link (id), Categories, SubmittedBy provided by user. Others defaults
		res, err := db.Exec("INSERT INTO Tags VALUES(?,?,?,?,?);", nil, tag_data.LinkID, tag_data.Categories, tag_data.SubmittedBy, tag_data.LastUpdated)
		if err != nil {
			log.Fatal(err)
		}

		var id int64
		if id, err = res.LastInsertId(); err != nil {
			log.Fatal(err)
		}
		tag_data.ID = id

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, tag_data)
	})

	// Serve
	// make sure this runs AFTER all routes
	if err := http.ListenAndServe("localhost:8000", r); err != nil {
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
	Created string
}

type User struct {
	*UserAuth
	ID int64
	About string
	ProfilePic string
}
type SignUpRequest struct {
	*UserAuth
	
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

// SUMMARY

type SummaryRequest struct {
	*SummaryCreateRequest
	*SummaryEditRequest
	*SummaryDeleteRequest
	*SummaryLikeRequest
	*SummaryLikeDeleteRequest
}

func (a *SummaryRequest) Bind(r *http.Request) error {
	if a.SummaryCreateRequest == nil && a.SummaryLikeRequest == nil && a.SummaryEditRequest == nil && a.SummaryDeleteRequest == nil && a.SummaryLikeDeleteRequest == nil {
		return errors.New("missing required Summary fields")
	}

	if a.SummaryEditRequest != nil {
		if a.SummaryEditRequest.Text == "" {
			return errors.New("missing replacement summary text")
		} else if a.SummaryEditRequest.SummaryID == "" {
			return errors.New("missing summary ID")
		}
	}

	return nil
}

type SummaryCreateRequest struct {
	SubmittedBy string `json:"submitted_by"`
	LinkID string `json:"link_id"`
	Text string `json:"text"`
}

type SummaryEditRequest struct {
	// would use json:"summary_id" here but conflicts with
	// below SummaryLikeRequest json ... not sure how else to fix
	SummaryID string `json:"summary_id_edit"`
	Text string `json:"text_edit"`
}

type SummaryDeleteRequest struct {
	// would use json:"summary_id" here but conflicts with
	// below SummaryLikeRequest json ... not sure how else to fix
	SummaryID string `json:"summary_id_del"`
}

type SummaryLikeRequest struct {
	SummaryID string `json:"summary_id"`
	UserID string `json:"user_id"`
}

type SummaryLikeDeleteRequest struct {
	SummaryLikeID string `json:"slike_id"`
}

// LINK
type Link struct {
	ID int64 `json:"link_id"`
	URL string `json:"url"`
	SubmittedBy string `json:"submitted_by"`
	SubmitDate string `json:"submit_date"`
}

type LinkRequest struct {
	*Link
}

func (a *LinkRequest) Bind(r *http.Request) error {
	if a.Link == nil {
		return errors.New("missing required Link fields")
	}

	a.SubmitDate = time.Now().Format("2006-01-02 15:04:05")

	return nil
}

// TAG
type Tag struct {
	ID int64 `json:"tag_id"`
	LinkID string `json:"link_id"`
	Categories string `json:"categories"`
	SubmittedBy string `json:"submitted_by"`
	LastUpdated string `json:"last_updated"`
}

type TagCreateRequest struct {
	*Tag
}

func (a *TagCreateRequest) Bind(r *http.Request) error {
	if a.Tag == nil {
		return errors.New("missing required Tag fields")
	}

	a.LastUpdated = time.Now().Format("2006-01-02 15:04:05")

	return nil
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