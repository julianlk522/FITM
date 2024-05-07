package handler

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/render"

	"oitm/model"
)

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

}