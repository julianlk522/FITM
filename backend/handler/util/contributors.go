package handler

import (
	"log"
	"net/http"

	"oitm/db"
	"oitm/model"
	"oitm/query"

	"github.com/go-chi/render"
)

// Contributors
func ScanContributors(contributors_sql *query.Contributors) *[]model.Contributor {
	rows, err := db.Client.Query(contributors_sql.Text)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	contributors := []model.Contributor{}
	for rows.Next() {
		contributor := model.Contributor{}
		err := rows.Scan(
			&contributor.LinksSubmitted,
			&contributor.LoginName,
		)
		if err != nil {
			log.Fatal(err)
		}
		contributors = append(contributors, contributor)
	}

	return &contributors
}

func RenderContributors(contributors *[]model.Contributor, w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, contributors)
}
