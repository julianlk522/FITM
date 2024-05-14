package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"oitm/model"
)

// GET SUMMARIES FOR LINK
func GetSummariesForLink(w http.ResponseWriter, r *http.Request) {

	link_id := chi.URLParam(r, "link_id")
	if link_id == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("no link id found")))
		return
	}

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// TODO: check auth token

	// Get summaries
	get_summaries_sql := fmt.Sprintf(`SELECT id, text, submitted_by FROM Summaries WHERE link_id = '%s'`, link_id)
	rows, err := db.Query(get_summaries_sql)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	summaries := []model.Summary{}

	if rows.Next() {
		for ok := true; ok; ok = rows.Next() {
			i := model.Summary{}
			err := rows.Scan(&i.ID, &i.Text, &i.SubmittedByID)
			if err != nil {
				log.Fatal(err)
			}
			summaries = append(summaries, i)
		
		}
	}

	render.JSON(w, r, summaries)
	render.Status(r, http.StatusOK)
}

// ADD / LIKE SUMMARY
// (depending on JSON fields supplied)
func AddSummaryOrSummaryLike(w http.ResponseWriter, r *http.Request) {
	summary_data := &model.SummaryRequest{}

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

		_, err := db.Exec(`INSERT INTO Summaries VALUES (?,?,?,?)`, nil, summary_data.NewSummaryRequest.Text, summary_data.NewSummaryRequest.LinkID, summary_data.NewSummaryRequest.SubmittedByID)
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

	// Recalculate global_summary
	recalc_global_summary(summary_data.LinkID, db)

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, summary_data)
}

// EDIT SUMMARY
func EditSummary(w http.ResponseWriter, r *http.Request) {
	edit_data := &model.SummaryRequest{}

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
}

// DELETE / UN-LIKE SUMMARY
// (depending on JSON fields supplied)
func DeleteOrUnlikeSummary(w http.ResponseWriter, r *http.Request) {
	summary_data := &model.SummaryRequest{}

	if err := render.Bind(r, summary_data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// TODO: check auth token

	// Declare link ID for summary / summary like to update global_summary later
	var lid sql.NullString
	var get_lid_sql string

	// Delete Summary
	if summary_data.DeleteSummaryRequest != nil {

		// Get link ID
		get_lid_sql = fmt.Sprintf(`SELECT link_id FROM Summaries WHERE id = '%s'`, summary_data.DeleteSummaryRequest.SummaryID)
		err = db.QueryRow(get_lid_sql).Scan(&lid)
		if err != nil {
			log.Fatal(err)
		}

		// Delete Summary
		_, err = db.Exec(`DELETE FROM Summaries WHERE id = ?`, summary_data.DeleteSummaryRequest.SummaryID)
		if err != nil {
			log.Fatal(err)
		}

	// Unlike Summary
	} else if summary_data.DeleteSummaryLikeRequest != nil {

		// Get link ID
		get_lid_sql = fmt.Sprintf(`SELECT link_id FROM Summaries WHERE Summaries.id IN (SELECT summary_id FROM 'Summary Likes' WHERE 'Summary Likes'.id = '%s');`, summary_data.DeleteSummaryLikeRequest.SummaryLikeID)
		err = db.QueryRow(get_lid_sql).Scan(&lid)
		if err != nil {
			log.Fatal(err)
		}

		// Delete Summary Like
		_, err = db.Exec(`DELETE FROM 'Summary Likes' WHERE id = ?`, summary_data.DeleteSummaryLikeRequest.SummaryLikeID)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Recalculate global_summary
	recalc_global_summary(lid.String, db)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{"status": "accepted"})

}

func recalc_global_summary(link_id string, db *sql.DB) {
	// Recalculate global_summary
	// (Summary with the most upvotes is the global summary)
	get_summary_like_counts_sql := fmt.Sprintf(`select text from summaries LEFT JOIN 'Summary Likes' ON summaries.id = 'Summary Likes'.summary_id WHERE link_id = '%s' GROUP BY summaries.id ORDER BY count(*) DESC, text ASC LIMIT 1;`, link_id)

	var top_summary_text string
	err := db.QueryRow(get_summary_like_counts_sql).Scan(&top_summary_text)
	if err != nil {
		log.Fatal(err)
	}

	// Update global_summary
	_, err = db.Exec(`UPDATE Links SET global_summary = ? WHERE id = ?`, top_summary_text, link_id)
	if err != nil {
		log.Fatal(err)
	}
}