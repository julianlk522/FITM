package handler

import (
	"log"
	"net/http"
	"os"

	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"

	"github.com/julianlk522/fitm/db"
	e "github.com/julianlk522/fitm/error"
	util "github.com/julianlk522/fitm/handler/util"
	m "github.com/julianlk522/fitm/middleware"
	"github.com/julianlk522/fitm/model"
	"github.com/julianlk522/fitm/query"
)

func GetLinks(w http.ResponseWriter, r *http.Request) {
	links_sql := query.NewTopLinks()

	// cats
	cats_params := r.URL.Query().Get("cats")
	if cats_params != "" {
		cats := strings.Split(cats_params, ",")
		links_sql = links_sql.FromCats(cats)
	}

	// period
	period_params := r.URL.Query().Get("period")
	if period_params != "" {
		links_sql = links_sql.DuringPeriod(period_params)
	}

	// auth fields
	req_user_id := r.Context().Value(m.UserIDKey).(string)
	if req_user_id != "" {
		links_sql = links_sql.AsSignedInUser(req_user_id)
	}

	// pagination
	page := r.Context().Value(m.PageKey).(int)
	links_sql = links_sql.Page(page)

	if links_sql.Error != nil {
		render.Render(w, r, e.ErrInvalidRequest(links_sql.Error))
		return
	}

	// scan
	if req_user_id != "" {
		links, err := util.ScanLinks[model.LinkSignedIn](links_sql)
		if err != nil {
			render.Render(w, r, e.ErrInvalidRequest(err))
		}
		render.JSON(w, r, util.PaginateLinks(links, page))
	} else {
		links, err := util.ScanLinks[model.Link](links_sql)
		if err != nil {
			render.Render(w, r, e.ErrInvalidRequest(err))
		}
		render.JSON(w, r, util.PaginateLinks(links, page))
	}
}

func AddLink(w http.ResponseWriter, r *http.Request) {
	request := &model.NewLinkRequest{}
	if err := render.Bind(r, request); err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	// Check if URL is YT video
	// if so, extract ID and send request to GoogleAPIs using API key
	if util.IsYouTubeVideoLink(request.NewLink.URL) {

		// Get ID from URL
		id := util.ExtractYouTubeVideoID(request.NewLink.URL)
		if id == "" {
			render.Render(w, r, e.ErrInvalidRequest(e.ErrInvalidURL))
			return
		}

		// Build new GoogleAPIs URL
		API_KEY := os.Getenv("FITM_GOOGLE_API_KEY")
		if API_KEY == "" {
			log.Printf("Could not find API_KEY")
			render.Status(r, http.StatusInternalServerError)
			return
		}
		gAPIs_url := "https://www.googleapis.com/youtube/v3/videos?id=" + id + "&key=" + API_KEY + "&part=snippet"

		// Request from GoogleAPIs
		resp, err := http.Get(gAPIs_url)
		if err != nil {
			log.Printf("Error requesting from GoogleAPIs: %s", err)
			render.Render(w, r, e.ErrInvalidRequest(err))
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Printf("Error response from GoogleAPIs: %s", err)
			render.Render(w, r, e.ErrInvalidRequest(e.ErrInvalidURL))
			return
		}

		// Extract URL metadata from response
		video_data, err := util.ExtractMetaDataFromGoogleAPIsResponse(resp.Body)
		if err != nil {
			log.Printf("Error: %s", err)
			render.Status(r, http.StatusInternalServerError)
			return
		}
		request.AutoSummary = video_data.Items[0].Snippet.Title
		request.ImgURL = video_data.Items[0].Snippet.Thumbnails.Default.URL
		request.URL = "https://www.youtube.com/watch?v=" + id

	} else {
		resp, err := util.ResolveURL(request.NewLink.URL)
		if err != nil {
			render.Render(w, r, e.ErrInvalidRequest(err))
			return
		}
		defer resp.Body.Close()

		// save updated URL (after any redirects e.g., to wwww.)
		// remove trailing slash
		request.URL = strings.TrimSuffix(resp.Request.URL.String(), "/")

		meta := util.GetMetaFromHTMLTokens(resp.Body)
		util.AssignMetadata(meta, request)
	}

	// Check URL is unique
	// Note: this comes after ResolveURL() because
	// the URL may be mutated slightly
	if util.URLAlreadyAdded(request.URL) {
		render.Render(w, r, e.ErrInvalidRequest(e.DuplicateURL(request.URL)))
		return
	}

	req_login_name := r.Context().Value(m.LoginNameKey).(string)
	request.SubmittedBy = req_login_name

	unsorted_cats := request.NewLink.Cats
	util.AssignSortedCats(unsorted_cats, request)

	// Insert link
	_, err := db.Client.Exec(
		"INSERT INTO Links VALUES(?,?,?,?,?,?,?);",
		request.ID,
		request.URL,
		request.SubmittedBy,
		request.SubmitDate,
		request.Cats,
		request.NewLink.Summary,
		request.ImgURL,
	)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	// Insert Summary(ies)
	// (might have user-submitted, Auto Summary, or both)
	if request.AutoSummary != "" {
		_, err = db.Client.Exec(
			"INSERT INTO Summaries VALUES(?,?,?,?,?);",
			uuid.New().String(),
			request.AutoSummary,
			request.ID,
			db.AUTO_SUMMARY_USER_ID,
			request.SubmitDate,
		)
		if err != nil {
			// continue... no auto summary
			// but log err
			log.Print("Error adding auto summary: ", err)
		} else {
			request.SummaryCount = 1
		}
	}

	req_user_id := r.Context().Value(m.UserIDKey).(string)
	if request.NewLink.Summary != "" {
		_, err = db.Client.Exec(
			"INSERT INTO Summaries VALUES(?,?,?,?,?);",
			uuid.New().String(),
			request.NewLink.Summary,
			request.ID,
			req_user_id,
			request.SubmitDate,
		)
		if err != nil {
			render.Render(w, r, e.ErrInvalidRequest(err))
			return
		} else {
			request.SummaryCount += 1
		}
	}

	// Insert tag
	_, err = db.Client.Exec(
		"INSERT INTO Tags VALUES(?,?,?,?,?);",
		uuid.New().String(),
		request.ID,
		request.Cats,
		request.SubmittedBy,
		request.SubmitDate,
	)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	// Return new link
	new_link := model.Link{
		ID:           request.ID,
		URL:          request.URL,
		SubmittedBy:  request.SubmittedBy,
		SubmitDate:   request.SubmitDate,
		Cats:         request.Cats,
		Summary:      request.NewLink.Summary,
		SummaryCount: request.SummaryCount,
		ImgURL:       request.ImgURL,
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, new_link)
}

func LikeLink(w http.ResponseWriter, r *http.Request) {
	link_id := chi.URLParam(r, "link_id")
	if link_id == "" {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoLinkID))
		return
	}

	req_login_name := r.Context().Value(m.LoginNameKey).(string)
	if util.UserSubmittedLink(req_login_name, link_id) {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrCannotLikeOwnLink))
		return
	}

	req_user_id := r.Context().Value(m.UserIDKey).(string)
	if util.UserHasLikedLink(req_user_id, link_id) {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrLinkAlreadyLiked))
		return
	}

	new_like_id := uuid.New().String()
	_, err := db.Client.Exec(
		"INSERT INTO 'Link Likes' VALUES(?,?,?);",
		new_like_id,
		link_id,
		req_user_id,
	)
	if err != nil {
		log.Fatal(err)
	}

	like_link_data := make(map[string]string, 1)
	like_link_data["ID"] = new_like_id

	render.Status(r, http.StatusOK)
	render.JSON(w, r, like_link_data)
}

func UnlikeLink(w http.ResponseWriter, r *http.Request) {
	link_id := chi.URLParam(r, "link_id")
	if link_id == "" {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoLinkID))
		return
	}

	req_user_id := r.Context().Value(m.UserIDKey).(string)
	if !util.UserHasLikedLink(req_user_id, link_id) {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrLinkNotLiked))
		return
	}

	_, err := db.Client.Exec(
		"DELETE FROM 'Link Likes' WHERE link_id = ? AND user_id = ?;",
		link_id,
		req_user_id,
	)
	if err != nil {
		log.Fatal(err)
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{"message": "deleted"})
}

func CopyLink(w http.ResponseWriter, r *http.Request) {
	link_id := chi.URLParam(r, "link_id")
	if link_id == "" {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoLinkID))
		return
	}

	req_login_name := r.Context().Value(m.LoginNameKey).(string)
	owns_link := util.UserSubmittedLink(req_login_name, link_id)
	if owns_link {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrCannotCopyOwnLink))
		return
	}

	req_user_id := r.Context().Value(m.UserIDKey).(string)
	already_copied := util.UserHasCopiedLink(req_user_id, link_id)
	if already_copied {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrLinkAlreadyCopied))
		return
	}

	new_copy_id := uuid.New().String()

	_, err := db.Client.Exec(
		"INSERT INTO 'Link Copies' VALUES(?,?,?);",
		new_copy_id,
		link_id,
		req_user_id,
	)
	if err != nil {
		log.Fatal(err)
	}

	return_json := map[string]string{
		"ID": new_copy_id,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, return_json)
}

func UncopyLink(w http.ResponseWriter, r *http.Request) {
	link_id := chi.URLParam(r, "link_id")
	if link_id == "" {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoLinkID))
		return
	}

	req_user_id := r.Context().Value(m.UserIDKey).(string)
	already_copied := util.UserHasCopiedLink(req_user_id, link_id)
	if !already_copied {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrLinkNotCopied))
		return
	}

	// Delete
	_, err := db.Client.Exec(
		"DELETE FROM 'Link Copies' WHERE link_id = ? AND user_id = ?;",
		link_id,
		req_user_id,
	)
	if err != nil {
		log.Fatal(err)
	}

	return_json := map[string]string{
		"message": "deleted",
	}

	render.JSON(w, r, return_json)
	render.Status(r, http.StatusNoContent)
}
