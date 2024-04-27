package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Home - check if server running
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World!")
	})

	// Link
	type Link struct {
	ID int `json:"link_id"`
	URL string `json:"url"`
	Submitted_By string `json:"submitted_by"`
	Submit_Date string `json:"submit_date"`
	Likes int `json:"likes"`
	Summaries string `json:"summaries"`
	}

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
		url, submitted_by := r.FormValue("url"), r.FormValue("submitted_by")
		fmt.Print(url, " ", submitted_by)
		if url == "" || submitted_by == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Missing URL or submitted_by\n")
			return
		}

		db, err := sql.Open("sqlite3", "./db/oitm.db")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		res, err := db.Exec("INSERT INTO Links VALUES(?,?);", url, submitted_by)
		if err != nil {
			log.Fatal(err)
		}

		var id int64
		if id, err = res.LastInsertId(); err != nil {
			log.Fatal(err)
		}

		fmt.Fprint(w, "Added link with ID: ", id)

	})

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

	if err := http.ListenAndServe("localhost:8000", r); err != nil {
		log.Fatal(err)
	}
}
