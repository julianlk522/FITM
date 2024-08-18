package handler

import (
	"log"
	"net/http"

	"oitm/db"
	"oitm/model"
	"oitm/query"

	"github.com/go-chi/render"
)

// Cats Contributors
func ScanCatsContributors(contributors_sql *query.CatsContributors, categories_str string) *[]model.CategoryContributor {
	rows, err := db.Client.Query(contributors_sql.Text)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	contributors := []model.CategoryContributor{}
	for rows.Next() {
		contributor := model.CategoryContributor{Categories: categories_str}
		err := rows.Scan(&contributor.LinksSubmitted, &contributor.LoginName)
		if err != nil {
			log.Fatal(err)
		}
		contributors = append(contributors, contributor)
	}

	return &contributors
}

func RenderCatsContributors(contributors *[]model.CategoryContributor, w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, contributors)
}
