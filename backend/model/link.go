package model

import (
	"database/sql"
	"errors"
	"log"
	"math"
	"net/http"
	"strings"
	"time"
)

type NewLink struct {
	URL string `json:"url"`
	Categories string `json:"categories"`
}

type NewLinkRequest struct {
	*NewLink
	ID int64
	SubmittedBy string
	SubmitDate string
	Summary string
	SummaryCount int
	LikeCount int64
	ImgURL string
}

func (a *NewLinkRequest) Bind(r *http.Request) error {
	if a.NewLink == nil {
		return errors.New("missing required Link fields")
	}

	a.SubmitDate = time.Now().Format("2006-01-02 15:04:05")
	a.LikeCount = 0

	return nil
}

type LinkCopyRequest struct {
	ID int64
	LinkID string `json:"link_id"`
}

func (a *LinkCopyRequest) Bind(r *http.Request) error {
	if a.LinkID == "" {
		return errors.New("missing link ID")
	}

	return nil
}

type DeleteLinkCopyRequest struct {
	ID string `json:"copy_id"`
}

func (a *DeleteLinkCopyRequest) Bind(r *http.Request) error {
	if a.ID == "" {
		return errors.New("missing copy ID")
	}

	return nil
}

type Link struct {
	ID int64
	URL string
	SubmittedBy string
	SubmitDate string
	Categories string
	Summary string
	SummaryCount int
	LikeCount int64
	ImgURL string
}

type LinkSignedIn struct {
	Link
	IsLiked bool
}

type CustomLinkCategories struct {
	LinkID int64
	Categories string
}

type CategoryContributor struct {
	Categories string
	LoginName string
	LinksSubmitted int
}

// Recalculate global categories for a link whose tags changed
func RecalcGlobalCats(db *sql.DB, link_id string) {
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