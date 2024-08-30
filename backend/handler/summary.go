package handler

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"

	"github.com/julianlk522/fitm/db"
	e "github.com/julianlk522/fitm/error"
	util "github.com/julianlk522/fitm/handler/util"
	m "github.com/julianlk522/fitm/middleware"
	"github.com/julianlk522/fitm/model"
)

func GetSummaryPage(w http.ResponseWriter, r *http.Request) {
	link_id := chi.URLParam(r, "link_id")
	if link_id == "" {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoLinkID))
		return
	}

	link_exists, err := util.LinkExists(link_id)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	} else if !link_exists {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoLinkWithID))
		return
	}

	summary_page, err := util.BuildSummaryPageForLink(link_id, r)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	render.JSON(w, r, summary_page)
}

func AddSummary(w http.ResponseWriter, r *http.Request) {
	summary_data := &model.NewSummaryRequest{}
	if err := render.Bind(r, summary_data); err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	req_user_id := r.Context().Value(m.UserIDKey).(string)
	link_exists, err := util.LinkExists(summary_data.LinkID)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	} else if !link_exists {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoLinkWithID))
		return
	}

	summary_id, err := util.GetIDOfUserSummaryForLink(req_user_id, summary_data.LinkID)
	if err != nil {
		if err == sql.ErrNoRows {

			// Create new summary
			_, err = db.Client.Exec(
				`INSERT INTO Summaries VALUES (?,?,?,?,?)`,
				summary_data.ID,
				summary_data.Text,
				summary_data.LinkID,
				req_user_id,
				summary_data.LastUpdated,
			)
			if err != nil {
				render.Render(w, r, e.ErrInvalidRequest(err))
				return
			}

		} else {
			render.Render(w, r, e.ErrInvalidRequest(err))
			return
		}

	} else {

		// Update summary if already submitted
		_, err = db.Client.Exec(
			`UPDATE Summaries SET text = ?, last_updated = ? WHERE submitted_by = ? AND link_id = ?`,
			summary_data.Text,
			summary_data.LastUpdated,
			req_user_id,
			summary_data.LinkID,
		)
		if err != nil {
			render.Render(w, r, e.ErrInvalidRequest(err))
			return
		}

		// Reset Summary Likes
		_, err = db.Client.Exec(
			`DELETE FROM 'Summary Likes' WHERE summary_id = ?`,
			summary_id,
		)
		if err != nil {
			render.Render(w, r, e.ErrInvalidRequest(err))
			return
		}
	}

	err = util.CalculateAndSetGlobalSummary(summary_data.LinkID)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, summary_data)
}

func DeleteSummary(w http.ResponseWriter, r *http.Request) {
	delete_data := &model.DeleteSummaryRequest{}
	if err := render.Bind(r, delete_data); err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	req_user_id := r.Context().Value(m.UserIDKey).(string)
	owns_summary, err := util.SummarySubmittedByUser(delete_data.SummaryID, req_user_id)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	} else if !owns_summary {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrDoesntOwnSummary))
		return
	}

	link_id, err := util.GetLinkIDFromSummaryID(delete_data.SummaryID)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	is_last_summary_for_link, err := util.LinkHasOneSummaryLeft(link_id)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	} else if is_last_summary_for_link {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrLastSummary))
		return
	}

	_, err = db.Client.Exec(
		`DELETE FROM Summaries WHERE id = ?`,
		delete_data.SummaryID,
	)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	err = util.CalculateAndSetGlobalSummary(link_id)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	util.RenderDeleted(w, r)
}

func LikeSummary(w http.ResponseWriter, r *http.Request) {
	summary_id := chi.URLParam(r, "summary_id")
	if summary_id == "" {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoSummaryID))
		return
	}

	var link_id sql.NullString
	err := db.Client.QueryRow(
		"SELECT link_id FROM Summaries WHERE id = ?",
		summary_id,
	).Scan(&link_id)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoSummaryWithID))
		return
	}

	req_login_name := r.Context().Value(m.LoginNameKey).(string)
	user_exists, err := util.UserExists(req_login_name)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
		} else if !user_exists {
			render.Render(w, r, e.ErrInvalidRequest(e.ErrNoUserWithLoginName))
		}
		
	req_user_id := r.Context().Value(m.UserIDKey).(string)
	owns_summary, err := util.SummarySubmittedByUser(summary_id, req_user_id)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	} else if owns_summary {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrCannotLikeOwnSummary))
		return
	}

	already_liked, err := util.UserHasLikedSummary(req_user_id, summary_id)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	} else if already_liked {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrSummaryAlreadyLiked))
		return
	}

	_, err = db.Client.Exec(
		`INSERT INTO 'Summary Likes' VALUES (?,?,?)`,
		uuid.New().String(),
		summary_id,
		req_user_id,
	)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	err = util.CalculateAndSetGlobalSummary(link_id.String)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{"message": "liked"})
}

func UnlikeSummary(w http.ResponseWriter, r *http.Request) {
	summary_id := chi.URLParam(r, "summary_id")
	if summary_id == "" {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoSummaryID))
		return
	}

	req_user_id := r.Context().Value(m.UserIDKey).(string)
	already_liked, err := util.UserHasLikedSummary(req_user_id, summary_id)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	} else if !already_liked {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrSummaryNotLiked))
		return
	}

	// save link id before delete
	// so global summary can be updated
	var link_id sql.NullString
	err = db.Client.QueryRow(
		"SELECT link_id FROM Summaries WHERE id = ?",
		summary_id,
	).Scan(&link_id)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoSummaryWithID))
		return
	}

	_, err = db.Client.Exec(
		`DELETE FROM 'Summary Likes' WHERE user_id = ? AND summary_id = ?`, req_user_id,
		summary_id,
	)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoSummaryWithID))
		return
	}

	err = util.CalculateAndSetGlobalSummary(link_id.String)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	util.RenderDeleted(w, r)
}
