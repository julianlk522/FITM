package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

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

	// Check if link exists, Abort if invalid link ID provided
	var l sql.NullString
	err = db.QueryRow("SELECT id FROM Links WHERE id = ?;", link_id).Scan(&l)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("no link found with given ID")))
		return
	}

	// Check auth token
	var req_user_id string
	claims, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if len(claims) > 0 {
		req_user_id = claims["user_id"].(string)
	}

	// authenticated
	if req_user_id != "" {

		// Get link
		get_link_sql := fmt.Sprintf(`SELECT links_id as link_id, url, submitted_by, submit_date, coalesce(categories,"") as categories, summary, COUNT('Link Likes'.id) as like_count, coalesce(is_liked,0) as is_liked, img_url
		FROM
			(
			SELECT id as links_id, url, submitted_by, submit_date, global_cats as categories, global_summary as summary, coalesce(img_url,"") as img_url
			FROM Links
			WHERE id = '%s'
			)
		LEFT JOIN 'Link Likes'
		ON 'Link Likes'.link_id = links_id
		LEFT JOIN
			(
			SELECT id, count(*) as is_liked, user_id, link_id as like_link_id2
			FROM 'Link Likes'
			WHERE user_id = '%s'
			GROUP BY id
			)
		ON like_link_id2 = link_id`, link_id, req_user_id)
		var link model.LinkSignedIn
		err = db.QueryRow(get_link_sql).Scan(&link.ID, &link.URL, &link.SubmittedBy, &link.SubmitDate, &link.Categories, &link.Summary, &link.LikeCount, &link.IsLiked, &link.ImgURL)
		if err != nil {
			if err == sql.ErrNoRows {
				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, ErrResponse{Err: errors.New("link not found")})
			} else {
				log.Fatal(err)
			}
		}

		// Get summaries and like counts
		get_summaries_sql := fmt.Sprintf(`SELECT sumid, text, login_name as submitted_by, coalesce(count(sl.id),0) as like_count, coalesce(is_liked,0) as is_liked
		FROM 
			(
			SELECT sumid, text, Users.login_name 
			FROM 
				(
				SELECT id as sumid, text, submitted_by 
				FROM Summaries 
				WHERE link_id = '%s'
				) 
			JOIN Users 
			ON Users.id = submitted_by
			)
		LEFT JOIN
			(
			SELECT id, count(*) as is_liked, user_id, summary_id as slsumid
			FROM 'Summary Likes'
			WHERE user_id = '%s'
			GROUP BY id
			)
		ON slsumid = sumid
		LEFT JOIN 'Summary Likes' as sl
		ON sl.summary_id = sumid 
		GROUP BY sumid;`, link_id, req_user_id)
		rows, err := db.Query(get_summaries_sql)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		summaries := []model.SummarySignedIn{}
		for rows.Next() {
			i := model.SummarySignedIn{}
			err := rows.Scan(&i.ID, &i.Text, &i.SubmittedBy, &i.LikeCount, &i.IsLiked)
			if err != nil {
				log.Fatal(err)
			}
			summaries = append(summaries, i)
		}
		if err != nil {
			log.Fatal(err)
		}

		summary_page := model.SummaryPageSignedIn{
			Link: link,
			Summaries: summaries,
		}

		render.JSON(w, r, summary_page)

	// unathenticated
	} else {

		// Get link
		get_link_sql := fmt.Sprintf(`SELECT links_id as link_id, url, submitted_by, submit_date, coalesce(categories,"") as categories, summary, COUNT('Link Likes'.id) as like_count, img_url FROM (SELECT id as links_id, url, submitted_by, submit_date, global_cats as categories, global_summary as summary, coalesce(img_url,"") as img_url FROM Links WHERE id = '%s') LEFT JOIN 'Link Likes' ON 'Link Likes'.link_id = links_id`, link_id)
		var link model.Link
		err = db.QueryRow(get_link_sql).Scan(&link.ID, &link.URL, &link.SubmittedBy, &link.SubmitDate, &link.Categories, &link.Summary, &link.LikeCount, &link.ImgURL)
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
	}
}

// ADD / LIKE SUMMARY
// (depending on JSON fields supplied)
func AddSummary(w http.ResponseWriter, r *http.Request) {
	summary_data := &model.NewSummaryRequest{}
	if err := render.Bind(r, summary_data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Check auth token
	var req_user_id string
	claims, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if len(claims) > 0 {
		req_user_id = claims["user_id"].(string)
	}

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

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

	_, err = db.Exec(`INSERT INTO Summaries VALUES (?,?,?,?)`, nil, summary_data.Text, summary_data.LinkID, req_user_id)
	if err != nil {
		log.Fatal(err)
	}

	// Recalculate global_summary
	RecalculateGlobalSummary(summary_data.LinkID, db)

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, summary_data)
}

// DELETE SUMMARY
func DeleteSummary(w http.ResponseWriter, r *http.Request) {
	delete_data := &model.DeleteSummaryRequest{}
	if err := render.Bind(r, delete_data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Check auth token
	var req_user_id string
	claims, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if len(claims) > 0 {
		req_user_id = claims["user_id"].(string)
	}
	
	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check that summary exists and submitted by user, Abort if not
	var u sql.NullInt64
	req_user_id_int64, err := strconv.ParseInt(req_user_id, 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	err = db.QueryRow("SELECT submitted_by FROM Summaries WHERE id = ?", delete_data.SummaryID).Scan(&u)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("summary not found")))
		return
	} else if u.Int64 != req_user_id_int64 {
		render.Render(w, r, ErrInvalidRequest(errors.New("not your summary")))
		return
	}
		
	// Get link ID
	var lid sql.NullString
	get_lid_sql := fmt.Sprintf(`SELECT link_id FROM Summaries WHERE id = '%s'`, delete_data.SummaryID)
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
	_, err = db.Exec(`DELETE FROM Summaries WHERE id = ?`, delete_data.SummaryID)
	if err != nil {
		log.Fatal(err)
	}

	// Recalculate global_summary
	RecalculateGlobalSummary(lid.String, db)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{"message": "deleted"})
}

// EDIT SUMMARY
func EditSummary(w http.ResponseWriter, r *http.Request) {
	edit_data := &model.EditSummaryRequest{}

	if err := render.Bind(r, edit_data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Check auth token
	var req_user_id string
	claims, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if len(claims) > 0 {
		req_user_id = claims["user_id"].(string)
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

	req_user_id_int64, err := strconv.ParseInt(req_user_id, 10, 64)
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

// LIKE SUMMARY
func LikeSummary(w http.ResponseWriter, r *http.Request) {
	summary_id := chi.URLParam(r, "summary_id")
	if summary_id == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("no summary ID provided")))
		return
	}

	// Check auth token
	var req_user_id string
	claims, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if len(claims) > 0 {
		req_user_id = claims["user_id"].(string)
	}
	
	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check if summary exists, Abort if not
	// Also save link_id for recalculating global_summary later
	var s, lid sql.NullString
	err = db.QueryRow("SELECT id, link_id FROM Summaries WHERE id = ?", summary_id).Scan(&s, &lid)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("summary not found")))
		return
	}

	// Check if user already liked summary, Abort if so
	var slid sql.NullString
	err = db.QueryRow("SELECT id FROM 'Summary Likes' WHERE summary_id = ? AND user_id = ?", summary_id, req_user_id).Scan(&slid)
	if err == nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("already liked")))
		return
	}

	// Add like
	_, err = db.Exec(`INSERT INTO 'Summary Likes' VALUES (?,?,?)`, nil, req_user_id, summary_id)
	if err != nil {
		log.Fatal(err)
	}

	// Recalculate global_summary
	RecalculateGlobalSummary(lid.String, db)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{"message": "liked"})
}

// DELETE / UN-LIKE SUMMARY
// (depending on JSON fields supplied)
func UnlikeSummary(w http.ResponseWriter, r *http.Request) {
	summary_id := chi.URLParam(r, "summary_id")
	if summary_id == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("no summary ID provided")))
		return
	}

	// Check auth token
	var req_user_id string
	claims, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if len(claims) > 0 {
		req_user_id = claims["user_id"].(string)
	}
	
	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check that user has liked summary with given ID, Abort if not
	var slid sql.NullString
	err = db.QueryRow("SELECT id FROM 'Summary Likes' WHERE summary_id = ? AND user_id = ?", summary_id, req_user_id).Scan(&slid)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("not liked")))
		return
	}

	// Get link ID before deleting
	var lid sql.NullString
	get_lid_sql := fmt.Sprintf(`SELECT link_id FROM Summaries WHERE Summaries.id = (SELECT summary_id FROM 'Summary Likes' WHERE 'Summary Likes'.id = '%s');`, slid.String)
	err = db.QueryRow(get_lid_sql).Scan(&lid)
	if err != nil {
		log.Fatal(err)
	}

	// Delete Summary Like
	_, err = db.Exec(`DELETE FROM 'Summary Likes' WHERE id = ?`, slid.String)
	if err != nil {
		log.Fatal(err)
	}

	// Recalculate global_summary
	RecalculateGlobalSummary(lid.String, db)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{"message": "deleted"})
}

func RecalculateGlobalSummary(link_id string, db *sql.DB) {
	
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