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
	
	*/
	
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Home - check if server running
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World!")
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
			err := rows.Scan(&i.ID, &i.URL, &i.Submitted_By, &i.Submit_Date, &i.Likes, &i.Summaries)
			if err != nil {
				log.Fatal(err)
			}
			links = append(links, i)
		}

		fmt.Fprint(w, links)
	})

	// Add new link
	r.Post("/links", func(w http.ResponseWriter, r *http.Request) {
		data := &LinkRequest{}
		if err := render.Bind(r, data); err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}

		db, err := sql.Open("sqlite3", "./db/oitm.db")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		url, submitted_by, submit_date, likes, summaries := data.URL, data.Submitted_By, data.Submit_Date, data.Likes, data.Summaries
		res, err := db.Exec("INSERT INTO Links VALUES(?,?,?,?,?,?);", nil, url, submitted_by, submit_date, likes, summaries)
		if err != nil {
			log.Fatal(err)
		}

		var id int64
		if id, err = res.LastInsertId(); err != nil {
			log.Fatal(err)
		}
		data.ID = id

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, data)

	})

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

// Types
type Link struct {
	ID int64 `json:"link_id"`
	URL string `json:"url"`
	Submitted_By string `json:"submitted_by"`
	Submit_Date string `json:"submit_date"`
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

	a.Likes = 0
	a.Submit_Date = time.Now().Format("2006-01-02 15:04:05")
	a.Summaries = "_"

	return nil
}

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