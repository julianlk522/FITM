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

	query "oitm/db/query"
	"oitm/model"
)

// GET SUMMARIES FOR LINK
func GetSummariesForLink(w http.ResponseWriter, r *http.Request) {
	link_id := chi.URLParam(r, "link_id")
	if link_id == "" {
		render.Render(w, r, ErrInvalidRequest(ErrNoLinkID))
		return
	}

	link_exists, err := _LinkExists(link_id)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if !link_exists {
		render.Render(w, r, ErrInvalidRequest(ErrNoLinkWithID))
		return
	}

	req_user_id, _, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if req_user_id != "" {
		summary_page, err := _GetSummaryPageSignedIn(link_id, req_user_id)
		if err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}

		render.JSON(w, r, summary_page)
	} else {
		summary_page, err := _GetSummaryPageSignedOut(link_id)
		if err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}

		render.JSON(w, r, summary_page)
	}

	render.Status(r, http.StatusOK)
}

func _GetSummaryPageSignedIn(link_id string, req_user_id string) (*model.SummaryPage[model.SummarySignedIn, model.LinkSignedIn], error) {

	// add Isliked / IsCopied to link query
	get_link_sql := query.NewGetSummaryPageLink(link_id).ForSignedInUser(req_user_id)
	if get_link_sql.Error != nil {
		return nil, get_link_sql.Error
	}

	var link model.LinkSignedIn
	err := DBClient.QueryRow(get_link_sql.Text).Scan(&link.ID, &link.URL, &link.SubmittedBy, &link.SubmitDate, &link.Categories, &link.Summary, &link.LikeCount, &link.ImgURL, &link.IsLiked, &link.IsCopied)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoLinkWithID
		} else {
			return nil, err
		}
	}

	// add IsLiked to summary query
	get_summaries_sql := query.NewGetSummaries(link_id).ForSignedInUser(req_user_id)
	if get_summaries_sql.Error != nil {
		return nil, get_summaries_sql.Error
	}

	rows, err := DBClient.Query(get_summaries_sql.Text)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := []model.SummarySignedIn{}
	for rows.Next() {
		i := model.SummarySignedIn{}
		err := rows.Scan(&i.ID, &i.Text, &i.SubmittedBy, &i.LastUpdated, &i.LikeCount, &i.IsLiked)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, i)
	}

	summary_page := model.SummaryPage[model.SummarySignedIn, model.LinkSignedIn] {
		Link: link,
		Summaries: summaries,
	}

	return &summary_page, nil
}

func _GetSummaryPageSignedOut(link_id string) (*model.SummaryPage[model.SummarySignedOut, model.LinkSignedOut], error) {
	get_link_sql := query.NewGetSummaryPageLink(link_id)
	if get_link_sql.Error != nil {
		return nil, get_link_sql.Error

	}

	var link model.LinkSignedOut
	err := DBClient.QueryRow(get_link_sql.Text).Scan(&link.ID, &link.URL, &link.SubmittedBy, &link.SubmitDate, &link.Categories, &link.Summary, &link.LikeCount, &link.ImgURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoLinkWithID
		} else {
			return nil, err
		}
	}

	get_summaries_sql := query.NewGetSummaries(link_id)
	if get_summaries_sql.Error != nil {
		return nil, get_summaries_sql.Error
	}

	rows, err := DBClient.Query(get_summaries_sql.Text)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := []model.SummarySignedOut{}
	for rows.Next() {
		i := model.SummarySignedOut{}
		err := rows.Scan(&i.ID, &i.Text, &i.SubmittedBy, &i.LastUpdated, &i.LikeCount)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, i)
	}

	summary_page := model.SummaryPage[model.SummarySignedOut, model.LinkSignedOut]{
		Link: link,
		Summaries: summaries,
	}

	return &summary_page, nil
}

// ADD SUMMARY OR REPLACE EXISTING
func AddSummary(w http.ResponseWriter, r *http.Request) {
	summary_data := &model.NewSummaryRequest{}
	if err := render.Bind(r, summary_data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	req_user_id, _, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	link_exists, err := _LinkExists(summary_data.LinkID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if !link_exists {
		render.Render(w, r, ErrInvalidRequest(ErrNoLinkWithID))
		return
	}

	summary_id, err := _GetSummaryIDForLink(req_user_id, summary_data.LinkID)
	if err != nil {
		if err == sql.ErrNoRows {

			// Create new summary
			_, err = DBClient.Exec(`INSERT INTO Summaries VALUES (?,?,?,?,?)`, nil, summary_data.Text, summary_data.LinkID, req_user_id, summary_data.LastUpdated)
			if err != nil {
				render.Render(w, r, ErrInvalidRequest(err))
				return
			}

		} else {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}

	} else {

		// Update summary if already submitted
		_, err = DBClient.Exec(`UPDATE Summaries SET text = ?, last_updated = ?
		WHERE submitted_by = ? AND link_id = ?`, summary_data.Text, summary_data.LastUpdated, req_user_id, summary_data.LinkID)
		if err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}

		// Reset Summary Likes
		_, err = DBClient.Exec(`DELETE FROM 'Summary Likes' WHERE summary_id = ?`, summary_id)
		if err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}
	}

	_RecalculateGlobalSummary(summary_data.LinkID)

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, summary_data)
}

func _LinkExists(link_id string) (bool, error) {
	var l sql.NullString
	err := DBClient.QueryRow("SELECT id FROM Links WHERE id = ?", link_id).Scan(&l)
	if err != nil {
		return false, err
	}

	return l.Valid, nil
}

func _GetSummaryIDForLink(user_id string, link_id string) (string, error) {
	var summary_id sql.NullString
	err := DBClient.QueryRow("SELECT id FROM Summaries WHERE submitted_by = ? AND link_id = ?", user_id, link_id).Scan(&summary_id)

	if err != nil {
		return "", err
	}

	return summary_id.String, nil
}

// DELETE SUMMARY
func DeleteSummary(w http.ResponseWriter, r *http.Request) {
	delete_data := &model.DeleteSummaryRequest{}
	if err := render.Bind(r, delete_data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	
	req_user_id, _, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	owns_summary, err := _SummarySubmittedByUser(delete_data.SummaryID, req_user_id)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if !owns_summary {
		render.Render(w, r, ErrInvalidRequest(errors.New("not your summary")))
		return
	}
		
	link_id, err := _GetLinkIDFromSummaryID(delete_data.SummaryID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	is_last_summary_for_link, err := _LinkHasOneSummaryLeft(link_id)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if is_last_summary_for_link {
		render.Render(w, r, ErrInvalidRequest(errors.New("last summary for link, cannot delete")))
		return
	}

	_, err = DBClient.Exec(`DELETE FROM Summaries WHERE id = ?`, delete_data.SummaryID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	_RecalculateGlobalSummary(link_id)
	_RenderDeleted(w, r)
}

func _GetLinkIDFromSummaryID(summary_id string) (string, error) {
	var lid sql.NullString
	get_lid_sql := fmt.Sprintf(`SELECT link_id FROM Summaries WHERE id = '%s'`, summary_id)
	err := DBClient.QueryRow(get_lid_sql).Scan(&lid)
	if err != nil {
		return "", err
	}

	return lid.String, nil
}

func _LinkHasOneSummaryLeft(link_id string) (bool, error) {
	var c sql.NullInt32
	err := DBClient.QueryRow("SELECT COUNT(id) FROM Summaries WHERE link_id = ?", link_id).Scan(&c)
	if err != nil {
		return false, err
	}

	return c.Int32 == 1, nil
}

// LIKE SUMMARY
func LikeSummary(w http.ResponseWriter, r *http.Request) {
	summary_id := chi.URLParam(r, "summary_id")
	if summary_id == "" {
		render.Render(w, r, ErrInvalidRequest(ErrNoSummaryID))
		return
	}

	req_user_id, _, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	
	var link_id sql.NullString
	err = DBClient.QueryRow("SELECT link_id FROM Summaries WHERE id = ?", summary_id).Scan(&link_id)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(ErrNoSummaryWithID))
		return
	}

	owns_summary, err := _SummarySubmittedByUser(summary_id, req_user_id)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if owns_summary {
		render.Render(w, r, ErrInvalidRequest(errors.New("cannot like your own summary")))
		return
	}

	already_liked, err := _UserHasLikedSummary(req_user_id, summary_id)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if already_liked {
		render.Render(w, r, ErrInvalidRequest(errors.New("already liked")))
		return
	}

	_, err = DBClient.Exec(`INSERT INTO 'Summary Likes' VALUES (?,?,?)`, nil, req_user_id, summary_id)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	_RecalculateGlobalSummary(link_id.String)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{"message": "liked"})
}

func _SummarySubmittedByUser(summary_id string, user_id string) (bool, error) {
	var submitted_by sql.NullInt64
	err := DBClient.QueryRow("SELECT submitted_by FROM Summaries WHERE id = ?", summary_id).Scan(&submitted_by)

	if err != nil {
		return false, err
	}

	return strconv.FormatInt(submitted_by.Int64, 10) == user_id, nil
}

// UNLIKE SUMMARY
func UnlikeSummary(w http.ResponseWriter, r *http.Request) {
	summary_id := chi.URLParam(r, "summary_id")
	if summary_id == "" {
		render.Render(w, r, ErrInvalidRequest(ErrNoSummaryID))
		return
	}

	req_user_id, _, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	
	already_liked, err := _UserHasLikedSummary(req_user_id, summary_id)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if !already_liked {
		render.Render(w, r, ErrInvalidRequest(errors.New("not liked")))
		return
	}

	// Get link ID (needed to update global summary after unlike)
	var link_id sql.NullString
	err = DBClient.QueryRow("SELECT link_id FROM Summaries WHERE id = ?", summary_id).Scan(&link_id)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(ErrNoSummaryWithID))
		return
	}

	_, err = DBClient.Exec(`DELETE FROM 'Summary Likes' WHERE user_id = ? AND summary_id = ?`, req_user_id, summary_id)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(ErrNoSummaryWithID))
		return
	}

	_RecalculateGlobalSummary(link_id.String)
	_RenderDeleted(w, r)
}

func _UserHasLikedSummary(user_id string, summary_id string) (bool, error) {
	var summary_like_id sql.NullString
	err := DBClient.QueryRow("SELECT id FROM 'Summary Likes' WHERE user_id = ? AND summary_id = ?", user_id, summary_id).Scan(&summary_like_id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return summary_like_id.Valid, nil
}

func _RecalculateGlobalSummary(link_id string) {

	// (Summary with the most upvotes is the global summary)
	get_summary_like_counts_sql := fmt.Sprintf(`SELECT text
	FROM Summaries
	LEFT JOIN
		(
		SELECT summary_id, count(*) as like_count
		FROM 'Summary Likes'
		GROUP BY summary_id
		)
	ON Summaries.id = summary_id
	WHERE link_id = '%s'
	GROUP BY Summaries.id
	ORDER BY like_count DESC, text ASC
	LIMIT 1
	`, link_id)
	var top_summary_text string
	err := DBClient.QueryRow(get_summary_like_counts_sql).Scan(&top_summary_text)
	if err != nil {
		log.Fatal(err)
	}

	// Update global_summary
	_, err = DBClient.Exec(`UPDATE Links SET global_summary = ? WHERE id = ?`, top_summary_text, link_id)
	if err != nil {
		log.Fatal(err)
	}
}

func _RenderDeleted(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{"message": "deleted"})
}