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
	/* API Actions
	
	USER ACCOUNTS:
	-Sign up
	-Log in
	-Update profile settings

	TREASURE MAPS:
	-Get user's own treasure map
	-Get global treasure map chunks
		-intersectional reports (popular, new, etc.)
		-sectional top rankings based on likes

	LINKS:
	-Add new link
	-Like existing link
	-Copy extisting link to user's treasure map
	-Remove link from user's treasure map

	SUMMARIES:
	-Like link summary
	-Submit alternative link summary

	TAGS:
	-Edit link tags
	-Add new tag category (done automatically when editing a link's tag to include a new category)
	
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

		// LoginName, Password provided by user
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

	// Login
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

	// Edit about
	r.Patch("/users", func(w http.ResponseWriter, r *http.Request) {
		edit_about_data := &EditAboutRequest{}

		if err := render.Bind(r, edit_about_data); err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}

		db, err := sql.Open("sqlite3", "./db/oitm.db")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		// TODO: check auth token

		// update Users table
		
		// TODO replace hard-coded id with id corresponding
		// to provided auth token
		_, err = db.Exec(`UPDATE Users SET about = ? WHERE id = ?`, edit_about_data.About, edit_about_data.AuthToken)
		if err != nil {
			log.Fatal(err)
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, edit_about_data.About)
	})

	// Get Links
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
			err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.Likes, &i.Summaries)
			if err != nil {
				log.Fatal(err)
			}
			links = append(links, i)
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, links)
	})

	// Add new link
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

		// URL, SubmittedBy provided by user. Others defaults
		res, err := db.Exec("INSERT INTO Links VALUES(?,?,?,?,?,?);", nil, link_data.URL, link_data.SubmittedBy, link_data.SubmitDate, link_data.Likes, link_data.Summaries)
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

	// Add new tag
	r.Post("/tags", func(w http.ResponseWriter, r *http.Request) {
		tag_data := &CreateTagRequest{}
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
		err = db.QueryRow("SELECT id FROM Links WHERE id = ?;", tag_data.Link).Scan(&s)
		if err != nil {
			render.Render(w, r, ErrInvalidRequest(errors.New("invalid link id provided")))
			return
		}

		// Check if duplicate (same link ID, submitted by), Abort if so
		err = db.QueryRow("SELECT id FROM Tags WHERE link_id = ? AND submitted_by = ?;", tag_data.Link, tag_data.SubmittedBy).Scan(&s)
		if err == nil {
			render.Render(w, r, ErrInvalidRequest(errors.New("duplicate tag")))
			return
		}

		// Insert new tag
		// Link (id), Categories, SubmittedBy provided by user. Others defaults
		res, err := db.Exec("INSERT INTO Tags VALUES(?,?,?,?,?);", nil, tag_data.Link, tag_data.Categories, tag_data.SubmittedBy, tag_data.LastUpdated)
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
	ID int64
	LoginName string `json:"login_name"`
	Password string `json:"password"`
	Created string
}

type User struct {
	*UserAuth
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

type EditAboutRequest struct {
	AuthToken string `json:"token"`
	About string `json:"about"`
}

func (a *EditAboutRequest) Bind(r *http.Request) error {
	if a.AuthToken == "" {
		return errors.New("missing auth token")
	}

	// TODO: check auth token

	return nil
}

// LINK
type Link struct {
	ID int64 `json:"link_id"`
	URL string `json:"url"`
	SubmittedBy string `json:"submitted_by"`
	SubmitDate string `json:"submit_date"`
	Likes int `json:"likes"`
	Summaries string `json:"summaries"`
}

type LinkRequest struct {
	*Link
}

func (a *LinkRequest) Bind(r *http.Request) error {
	if a.Link == nil {
		return errors.New("missing required Link fields")
	}

	a.Likes = 0 // soon to be changed when Links db implemented
	a.SubmitDate = time.Now().Format("2006-01-02 15:04:05")
	a.Summaries = "_" // soon to be changed when Summaries db implemented

	return nil
}

// TAG
type Tag struct {
	ID int64 `json:"tag_id"`
	Link string `json:"link_id"`
	Categories string `json:"categories"`
	SubmittedBy string `json:"submitted_by"`
	LastUpdated string `json:"last_updated"`
}

type CreateTagRequest struct {
	*Tag
}

func (a *CreateTagRequest) Bind(r *http.Request) error {
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