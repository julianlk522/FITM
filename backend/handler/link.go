package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"

	"oitm/db"
	e "oitm/error"
	util "oitm/handler/util"
	m "oitm/middleware"
	"oitm/model"
	"oitm/query"
)

func GetLinks(w http.ResponseWriter, r *http.Request) {
	links_sql := query.NewTopLinks()

	cats_params := r.URL.Query().Get("cats")
	if cats_params != "" {
		link_ids, err := util.GetIDsOfLinksHavingCategories(cats_params)
		if err != nil {
			render.Render(w, r, e.ErrInvalidRequest(err))
			return
		} else if len(link_ids) == 0 {
			util.RenderZeroLinks(w, r)
			return
		}

		links_sql = links_sql.FromLinkIDs(link_ids)
	}

	period_params := r.URL.Query().Get("period")
	if period_params != "" {
		links_sql = links_sql.DuringPeriod(period_params)
	}

	page := r.Context().Value(m.PageKey).(int)
	links_sql = links_sql.Page(page)

	if links_sql.Error != nil {
		render.Render(w, r, e.ErrInvalidRequest(links_sql.Error))
		return
	}
	
	req_user_id := r.Context().Value(m.UserIDKey).(string)
	if req_user_id != "" {
		links, err := util.ScanLinks[model.LinkSignedIn](links_sql, req_user_id)
		if err != nil {
			render.Render(w, r, e.ErrInvalidRequest(err))
			return
		}
		util.RenderPaginatedLinks(links, page, w, r)
	} else {
		links, err := util.ScanLinks[model.Link](links_sql, req_user_id)
		if err != nil {
			render.Render(w, r, e.ErrInvalidRequest(err))
			return
		}
		util.RenderPaginatedLinks(links, page, w, r)
	}

}

func GetCatsContributors(w http.ResponseWriter, r *http.Request) {
	cats_params := chi.URLParam(r, "cats")
	if cats_params == "" {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoCats))
		return
	}
	cats := strings.Split(cats_params, ",")
	contributors_sql := query.NewCatsContributors(cats)

	period_params := r.URL.Query().Get("period")
	if period_params != "" {
		contributors_sql = contributors_sql.DuringPeriod(period_params)
	}

	if contributors_sql.Error != nil {
		render.Render(w, r, e.ErrInvalidRequest(contributors_sql.Error))
		return
	}
	
	contributors := util.ScanCatsContributors(contributors_sql, cats_params)
	util.RenderCatsContributors(contributors, w, r)
}

func GetSubcats(w http.ResponseWriter, r *http.Request) {
	cats_params := chi.URLParam(r, "cats")
	if cats_params == "" {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoCats))
		return
	}

	// TODO: replace with middleware that converts all URLs to lowercase
	// maybe encode uppercase chars another way?
	// TODO: figure out how other sites do that
	cats_params = strings.ToLower(cats_params)
	categories := strings.Split(cats_params, ",")
	subcats_sql := query.NewSubcats(categories)
	
	period_params := r.URL.Query().Get("period")
	if period_params != "" {
		subcats_sql = subcats_sql.DuringPeriod(period_params)
	}
	
	if subcats_sql.Error != nil {
		render.Render(w, r, e.ErrInvalidRequest(subcats_sql.Error))
		return
	}

	subcats := util.ScanSubcats(subcats_sql, categories)
	if len(subcats) == 0 {
		util.RenderZeroSubcategories(w, r)
		return
	}
	util.RenderSubcategories(subcats, categories, w, r)
}



func AddLink(w http.ResponseWriter, r *http.Request) {
	request := &model.NewLinkRequest{}
	if err := render.Bind(r, request); err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	// Check URL is valid
    resp, err := util.ResolveAndAssignURL(request.NewLink.URL, request)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}
	defer resp.Body.Close()

	// Check URL is unique
	// Note: this comes after ResolveAndAssignURL() because
	// the URL may be mutated, e.g., add www.
	if util.URLAlreadyAdded(request.URL) {
		render.Render(w, r, e.ErrInvalidRequest(fmt.Errorf("duplicate URL: %s", request.URL)))
		return
	}

	req_login_name := r.Context().Value(m.LoginNameKey).(string)
	request.SubmittedBy = req_login_name
	
	meta := util.MetaFromHTMLTokens(resp.Body)
	util.AssignMetadata(meta, request)

	unsorted_cats := request.NewLink.Categories
	util.AssignSortedCategories(unsorted_cats, request)
	
	// Insert link
	res, err := db.Client.Exec(
		"INSERT INTO Links VALUES(?,?,?,?,?,?,?);", 
		nil, 
		request.URL, 
		req_login_name, 
		request.SubmitDate, 
		request.Categories, 
		request.NewLink.Summary, 
		request.ImgURL,
	)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}
	
	if err := util.AssignNewLinkID(res, request); err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	// Insert Summary(ies)
	// (might have user-submitted, Auto Summary, or both)
	if request.AutoSummary != "" {
		// Note: UserID 15 is AutoSummary
		// TODO: add constant, replace magic 15
		_, err = db.Client.Exec(
			"INSERT INTO Summaries VALUES(?,?,?,?,?);", 
			nil, 
			request.AutoSummary, 
			request.ID, 
			"15", 
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
			nil, 
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
		request.Categories, 
		req_login_name, 
		request.SubmitDate,
	)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	// Return new link
	new_link := model.Link{
		ID: request.ID,
		URL: request.URL,
		SubmittedBy: req_login_name,
		SubmitDate: request.SubmitDate,
		Categories: request.Categories,
		Summary: request.NewLink.Summary,
		SummaryCount: request.SummaryCount,
		ImgURL: request.ImgURL,
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
		render.Render(w, r, e.ErrInvalidRequest(errors.New("cannot like your own link")))
		return
	}

	req_user_id := r.Context().Value(m.UserIDKey).(string)
	if util.UserHasLikedLink(req_user_id, link_id) {
		render.Render(w, r, e.ErrInvalidRequest(errors.New("already liked")))
		return
	}

	res, err := db.Client.Exec("INSERT INTO 'Link Likes' VALUES(?,?,?);", nil, req_user_id, link_id)
	if err != nil {
		log.Fatal(err)
	}

	var id int64
	if id, err = res.LastInsertId(); err != nil {
		log.Fatal(err)
	}

	like_link_data := make(map[string]int64, 1)
	like_link_data["ID"] = id
	
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
		render.Render(w, r, e.ErrInvalidRequest(errors.New("link like not found")))
		return
	}

	_, err := db.Client.Exec("DELETE FROM 'Link Likes' WHERE user_id = ? AND link_id = ?;", req_user_id, link_id)
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
		render.Render(w, r, e.ErrInvalidRequest(errors.New("cannot copy your own link to your treasure map")))
		return
	}

	req_user_id := r.Context().Value(m.UserIDKey).(string)
	already_copied := util.UserHasCopiedLink(req_user_id, link_id)
	if already_copied {
		render.Render(w, r, e.ErrInvalidRequest(errors.New("link already copied to treasure map")))
		return
	}

	res, err := db.Client.Exec("INSERT INTO 'Link Copies' VALUES(?,?,?);", nil, link_id, req_user_id)
	if err != nil {
		log.Fatal(err)
	}

	var id int64
	if id, err = res.LastInsertId(); err != nil {
		log.Fatal(err)
	}

	return_json := map[string]int64{
		"ID": id,
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
		render.Render(w, r, e.ErrInvalidRequest(errors.New("link copy does not exist")))
		return
	}

	// Delete
	_, err := db.Client.Exec("DELETE FROM 'Link Copies' WHERE user_id = ? AND link_id = ?;", req_user_id, link_id)
	if err != nil {
		log.Fatal(err)
	}

	return_json := map[string]string{
		"message": "deleted",
	}

	render.JSON(w, r, return_json)
	render.Status(r, http.StatusNoContent)
}