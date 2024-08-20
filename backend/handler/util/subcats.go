package handler

import (
	"log"
	"net/http"
	"strings"

	"oitm/db"
	e "oitm/error"
	"oitm/model"
	"oitm/query"

	"github.com/go-chi/render"
	"golang.org/x/exp/slices"
)

// Subcats
func ScanSubcats(get_subcats_sql *query.Subcats, search_cats []string) []string {
	rows, err := db.Client.Query(get_subcats_sql.Text)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var subcats []string
	for rows.Next() {
		var row_cats string
		if err := rows.Scan(&row_cats); err != nil {
			log.Fatal(err)
		}

		cats := strings.Split(row_cats, ",")
		for i := 0; i < len(cats); i++ {
			if !slices.Contains(search_cats, cats[i]) && !slices.Contains(subcats, cats[i]) {
				subcats = append(subcats, cats[i])
			}
		}
	}

	return subcats
}

func RenderZeroSubcategories(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, []model.CatCount{})
	render.Status(r, http.StatusOK)
}

func RenderSubcategories(subcats []string, categories []string, w http.ResponseWriter, r *http.Request) {
	with_counts, err := GetCountsOfSubcatsFromCats(subcats, categories)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	SortAndLimitCatCounts(with_counts, CATEGORY_PAGE_LIMIT)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, with_counts)
}

func GetCountsOfSubcatsFromCats(subcats []string, cats []string) (*[]model.CatCount, error) {
	subcat_counts := make([]model.CatCount, len(subcats))
	for i := 0; i < len(subcats); i++ {
		subcat_counts[i].Category = subcats[i]

		all_cats := append(cats, subcats[i])
		get_link_count_sql := query.NewCatCount(all_cats)
		if get_link_count_sql.Error != nil {
			return nil, get_link_count_sql.Error
		}

		if err := db.Client.QueryRow(get_link_count_sql.Text).Scan(&subcat_counts[i].Count); err != nil {
			return nil, err
		}
	}

	return &subcat_counts, nil
}

func SortAndLimitCatCounts(cat_counts *[]model.CatCount, limit int) {
	slices.SortFunc(*cat_counts, model.SortCategories)

	if len(*cat_counts) > limit {
		*cat_counts = (*cat_counts)[:limit]
	}
}
