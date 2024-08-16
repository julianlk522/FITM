package handler

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"net/http"

	"github.com/go-chi/render"

	"oitm/db"
	e "oitm/error"
	"oitm/model"
	"oitm/query"
)

// Get summaries
func GetSummaryPageSignedIn(link_id string, req_user_id string) (*model.SummaryPage[model.SummarySignedIn, model.LinkSignedIn], error) {

	// add Isliked / IsCopied to link query
	get_link_sql := query.
		NewSummaryPageLink(link_id).
		ForSignedInUser(req_user_id)
	if get_link_sql.Error != nil {
		return nil, get_link_sql.Error
	}

	var l model.LinkSignedIn
	err := db.Client.QueryRow(get_link_sql.Text).Scan(
		&l.ID, 
		&l.URL, 
		&l.SubmittedBy, 
		&l.SubmitDate, 
		&l.Categories, 
		&l.Summary, 
		&l.LikeCount, 
		&l.TagCount,
		&l.ImgURL, 
		&l.IsLiked, 
		&l.IsCopied,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, e.ErrNoLinkWithID
		} else {
			return nil, err
		}
	}

	// add IsLiked to summary query
	get_summaries_sql := query.
		NewSummariesForLink(link_id).
		ForSignedInUser(req_user_id)
	if get_summaries_sql.Error != nil {
		return nil, get_summaries_sql.Error
	}

	rows, err := db.Client.Query(get_summaries_sql.Text)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := []model.SummarySignedIn{}
	for rows.Next() {
		s := model.SummarySignedIn{}
		err := rows.Scan(
			&s.ID, 
			&s.Text, 
			&s.SubmittedBy, 
			&s.LastUpdated, 
			&s.LikeCount, 
			&s.IsLiked,
		)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}

	summary_page := model.SummaryPage[model.SummarySignedIn, model.LinkSignedIn] {
		Link: l,
		Summaries: summaries,
	}

	return &summary_page, nil
}

func GetSummaryPage(link_id string) (*model.SummaryPage[model.Summary, model.Link], error) {
	get_link_sql := query.NewSummaryPageLink(link_id)
	if get_link_sql.Error != nil {
		return nil, get_link_sql.Error

	}

	var l model.Link
	err := db.Client.QueryRow(get_link_sql.Text).Scan(
		&l.ID, 
		&l.URL, 
		&l.SubmittedBy, 
		&l.SubmitDate, 
		&l.Categories, 
		&l.Summary, 
		&l.LikeCount, 
		&l.TagCount,
		&l.ImgURL,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, e.ErrNoLinkWithID
		} else {
			return nil, err
		}
	}

	get_summaries_sql := query.NewSummariesForLink(link_id)
	if get_summaries_sql.Error != nil {
		return nil, get_summaries_sql.Error
	}

	rows, err := db.Client.Query(get_summaries_sql.Text)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := []model.Summary{}
	for rows.Next() {
		s := model.Summary{}
		err := rows.Scan(
			&s.ID, 
			&s.Text, 
			&s.SubmittedBy, 
			&s.LastUpdated, 
			&s.LikeCount,
		)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}

	summary_page := model.SummaryPage[model.Summary, model.Link]{
		Link: l,
		Summaries: summaries,
	}

	return &summary_page, nil
}



// Add summary
func LinkExists(link_id string) (bool, error) {
	var l sql.NullString
	err := db.Client.QueryRow("SELECT id FROM Links WHERE id = ?", link_id).Scan(&l)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	return l.Valid, nil
}

func GetIDOfUserSummaryForLink(user_id string, link_id string) (string, error) {
	var summary_id sql.NullString
	err := db.Client.QueryRow("SELECT id FROM Summaries WHERE submitted_by = ? AND link_id = ?", user_id, link_id).Scan(&summary_id)

	if err != nil {
		return "", err
	}

	return summary_id.String, nil
}



// Delete summary
func GetLinkIDFromSummaryID(summary_id string) (string, error) {
	var lid sql.NullString
	get_lid_sql := fmt.Sprintf(`SELECT link_id FROM Summaries WHERE id = '%s'`, summary_id)
	err := db.Client.QueryRow(get_lid_sql).Scan(&lid)
	if err != nil {
		return "", err
	}

	return lid.String, nil
}

func LinkHasOneSummaryLeft(link_id string) (bool, error) {
	var c sql.NullInt32
	err := db.Client.QueryRow("SELECT COUNT(id) FROM Summaries WHERE link_id = ?", link_id).Scan(&c)
	if err != nil {
		return false, err
	}

	return c.Int32 == 1, nil
}



// Like / unlike summary
func SummarySubmittedByUser(summary_id string, user_id string) (bool, error) {
	var submitted_by sql.NullInt64
	err := db.Client.QueryRow("SELECT submitted_by FROM Summaries WHERE id = ?", summary_id).Scan(&submitted_by)

	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	return strconv.FormatInt(submitted_by.Int64, 10) == user_id, nil
}

func UserHasLikedSummary(user_id string, summary_id string) (bool, error) {
	var summary_like_id sql.NullString
	err := db.Client.QueryRow("SELECT id FROM 'Summary Likes' WHERE user_id = ? AND summary_id = ?", user_id, summary_id).Scan(&summary_like_id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return summary_like_id.Valid, nil
}

func RenderDeleted(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{"message": "deleted"})
}



// Calculate global summary
func CalculateAndSetGlobalSummary(link_id string) error {

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

	err := db.Client.QueryRow(get_summary_like_counts_sql).Scan(&top_summary_text)
	if err != nil {
		return err
	}

	// Set global summary if not already set to query result
	check_global_summary_sql := fmt.Sprintf(`
		SELECT global_summary 
		FROM Links 
		WHERE id = '%s'`, 
	link_id)
	var gs string

	err = db.Client.QueryRow(check_global_summary_sql).Scan(&gs)
	if err != nil {
		return err
	} else if gs == "" || gs != top_summary_text {
		SetLinkGlobalSummary(link_id, top_summary_text)
	}

	return nil
}

func SetLinkGlobalSummary(link_id string, text string) {
	_, err := db.Client.Exec(`UPDATE Links SET global_summary = ? WHERE id = ?`, text, link_id)
	if err != nil {
		log.Fatal(err)
	}
}