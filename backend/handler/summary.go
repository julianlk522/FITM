package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
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

	// Check if link exists, Abort if invalid link ID provided
	var l sql.NullString
	err = db.QueryRow("SELECT id FROM Links WHERE id = ?;", link_id).Scan(&l)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("no link found with given ID")))
		return
	}

	// Get link
	get_link_sql := fmt.Sprintf(`SELECT links_id as link_id, url, submitted_by, submit_date, coalesce(categories,"") as categories, summary, COUNT('Link Likes'.id) as like_count FROM (SELECT id as links_id, url, submitted_by, submit_date, global_cats as categories, global_summary as summary FROM Links WHERE id = '%s') LEFT JOIN 'Link Likes' ON 'Link Likes'.link_id = links_id`, link_id)
	var link model.Link
	err = db.QueryRow(get_link_sql).Scan(&link.ID, &link.URL, &link.SubmittedBy, &link.SubmitDate, &link.Categories, &link.Summary, &link.LikeCount)
	if err != nil {
		if err == sql.ErrNoRows {
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, ErrResponse{Err: errors.New("link not found")})
		} else {
			log.Fatal(err)
		}
	}

	// Get summaries and like counts
	get_summaries_sql := fmt.Sprintf(`SELECT sumid, text, login_name, coalesce(count('Summary Likes'.id),0) as like_count FROM (SELECT sumid, text, Users.login_name FROM (SELECT id as sumid, text, submitted_by FROM Summaries WHERE link_id = '%s') JOIN Users ON Users.id = submitted_by) LEFT JOIN 'Summary Likes' ON 'Summary Likes'.summary_id = sumid GROUP BY sumid;`, link_id)
	rows, err := db.Query(get_summaries_sql)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	summaries := []model.Summary{}
	for rows.Next() {
		i := model.Summary{}
		err := rows.Scan(&i.ID, &i.Text, &i.SubmittedBy, &i.LikeCount)
		if err != nil {
			log.Fatal(err)
		}
		summaries = append(summaries, i)
	}
	if err != nil {
		log.Fatal(err)
	}

	summary_page := model.SummaryPage{
		Link: link,
		Summaries: summaries,
	}

	render.JSON(w, r, summary_page)
	render.Status(r, http.StatusOK)
}

// ADD / LIKE SUMMARY
// (depending on JSON fields supplied)
func AddSummaryOrSummaryLike(w http.ResponseWriter, r *http.Request) {
	summary_data := &model.NewSummaryOrSummaryLikeRequest{}

	if err := render.Bind(r, summary_data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Check auth token
	_, claims, err := jwtauth.FromContext(r.Context())
	// claims = {"user_id":"1234","login_name":"johndoe"}
	if err != nil {
		log.Fatal(err)
	}
	req_user_id, ok := claims["user_id"]
	if !ok {
		log.Fatal("invalid auth token")
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

		// Check if user already submitted a summary to this link, Abort if so
		var lid sql.NullString
		err = db.QueryRow("SELECT id FROM Summaries WHERE link_id = ? AND submitted_by = ?", summary_data.LinkID, req_user_id).Scan(&lid)
		if err == nil {
			render.Render(w, r, ErrInvalidRequest(errors.New("existing summary found from user for link")))
			return
		}

		_, err := db.Exec(`INSERT INTO Summaries VALUES (?,?,?,?)`, nil, summary_data.NewSummaryRequest.Text, summary_data.NewSummaryRequest.LinkID, req_user_id)
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

		// Check if user already liked summary, Abort if so
		var lid sql.NullString
		err = db.QueryRow("SELECT id FROM 'Summary Likes' WHERE summary_id = ? AND user_id = ?", summary_data.NewSummaryLikeRequest.SummaryID, req_user_id).Scan(&lid)
		if err == nil {
			render.Render(w, r, ErrInvalidRequest(errors.New("already liked")))
			return
		}

		_, err = db.Exec(`INSERT INTO 'Summary Likes' VALUES (?,?,?)`, nil, req_user_id, summary_data.NewSummaryLikeRequest.SummaryID)
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
	edit_data := &model.EditSummaryRequest{}

	if err := render.Bind(r, edit_data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Check auth token
	_, claims, err := jwtauth.FromContext(r.Context())
	// claims = {"user_id":"1234","login_name":"johndoe"}
	if err != nil {
		log.Fatal(err)
	}
	req_user_id, ok := claims["user_id"]
	if !ok {
		log.Fatal("invalid auth token")
	}

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check if summary doesn't exist or submitted by a different user, Abort if either
	var s sql.NullString
	var u sql.NullInt64
	err = db.QueryRow("SELECT id, submitted_by FROM Summaries WHERE id = ?", edit_data.SummaryID).Scan(&s, &u)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("summary not found")))
		return
	}

	req_user_id_int64, err := strconv.ParseInt(req_user_id.(string), 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	if u.Int64 != req_user_id_int64 {
		render.Render(w, r, ErrInvalidRequest(errors.New("cannot edit another user's summary")))
		return
	}

	// Update summary
	_, err = db.Exec(`UPDATE Summaries SET text = ? WHERE id = ?`, edit_data.Text, edit_data.SummaryID)
	if err != nil {
		log.Fatal(err)
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, edit_data)
}

// DELETE / UN-LIKE SUMMARY
// (depending on JSON fields supplied)
func DeleteOrUnlikeSummary(w http.ResponseWriter, r *http.Request) {
	summary_data := &model.DeleteSummaryOrSummaryLikeRequest{}
	if err := render.Bind(r, summary_data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Check auth token
	_, claims, err := jwtauth.FromContext(r.Context())
	// claims = {"user_id":"1234","login_name":"johndoe"}
	if err != nil {
		log.Fatal(err)
	}
	req_user_id, ok := claims["user_id"]
	if !ok {
		log.Fatal("invalid auth token")
	}

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// req_user_id_int64 will be used to check if user submitted summary / summary like
	req_user_id_int64, err := strconv.ParseInt(req_user_id.(string), 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	// link ID will be used to recalculate link's new global_summary after summary deletion
	var lid sql.NullString
	var get_lid_sql string

	// Delete Summary
	if summary_data.DeleteSummaryRequest != nil {

		// Check that summary exists and submitted by user, Abort if not
		var u sql.NullInt64
		err = db.QueryRow("SELECT submitted_by FROM Summaries WHERE id = ?", summary_data.DeleteSummaryRequest.SummaryID).Scan(&u)
		if err != nil {
			render.Render(w, r, ErrInvalidRequest(errors.New("summary not found")))
			return
		} else if u.Int64 != req_user_id_int64 {
			render.Render(w, r, ErrInvalidRequest(errors.New("not your summary")))
			return
		}

		// Get link ID
		get_lid_sql = fmt.Sprintf(`SELECT link_id FROM Summaries WHERE id = '%s'`, summary_data.DeleteSummaryRequest.SummaryID)
		err = db.QueryRow(get_lid_sql).Scan(&lid)
		if err != nil {
			log.Fatal(err)
		}

		// Check that summary is not only summary for its link, Abort if so
		var c sql.NullInt32
		err = db.QueryRow("SELECT COUNT(id) FROM Summaries WHERE link_id = ?", lid.String).Scan(&c)
		if err != nil {
			log.Fatal(err)
		} else if c.Int32 == 1 {
			render.Render(w, r, ErrInvalidRequest(errors.New("last summary for link, cannot delete")))
			return
		}

		// Delete Summary
		_, err = db.Exec(`DELETE FROM Summaries WHERE id = ?`, summary_data.DeleteSummaryRequest.SummaryID)
		if err != nil {
			log.Fatal(err)
		}

	// Unlike Summary
	} else if summary_data.DeleteSummaryLikeRequest != nil {

		// Check that summary like exists and submitted by user, Abort if not
		var u sql.NullInt64
		err = db.QueryRow("SELECT user_id FROM 'Summary Likes' WHERE id = ?", summary_data.DeleteSummaryLikeRequest.SummaryLikeID).Scan(&u)
		if err != nil {
			render.Render(w, r, ErrInvalidRequest(errors.New("summary like not found")))
			return
		} else if u.Int64 != req_user_id_int64 {
			render.Render(w, r, ErrInvalidRequest(errors.New("not your summary like")))
			return
		}

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
	render.JSON(w, r, map[string]string{"message": "deleted"})

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