package handler

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"oitm/db"
	e "oitm/error"
	util "oitm/handler/util"
	m "oitm/middleware"
	"oitm/model"
	"oitm/query"
)

func GetTagPage(w http.ResponseWriter, r *http.Request) {
	link_id := chi.URLParam(r, "link_id")
	if link_id == "" {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoLinkID))
		return
	}

	link_exists, err := util.LinkExists(link_id)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}
	if !link_exists {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoLinkWithID))
		return
	}

	req_user_id := r.Context().Value(m.UserIDKey).(string)
	link_sql := query.NewTagPageLink(link_id, req_user_id)
	if link_sql.Error != nil {
		render.Render(w, r, e.ErrInvalidRequest(link_sql.Error))
		return
	}

	link, err := util.ScanTagPageLink(link_sql)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	req_login_name := r.Context().Value(m.LoginNameKey).(string)
	user_tag, err := util.GetUserTagForLink(req_login_name, link_id)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	tag_rankings_sql := query.NewTagRankingsForLink(link_id)
	if tag_rankings_sql.Error != nil {
		render.Render(w, r, e.ErrInvalidRequest(tag_rankings_sql.Error))
		return
	}

	tag_rankings, err := util.ScanTagRankings(tag_rankings_sql)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	tag_page := model.TagPage{
		Link: link,
		UserTag: user_tag,
		TagRankings: tag_rankings,
	}
	render.JSON(w, r, tag_page)
}

func GetTopGlobalCats(w http.ResponseWriter, r *http.Request) {	
	global_cats_sql := query.NewAllGlobalCats()

	period_params := r.URL.Query().Get("period")
	has_period := period_params != ""
	if has_period {
		global_cats_sql = global_cats_sql.DuringPeriod(period_params)
	}

	if global_cats_sql.Error != nil {
		render.Render(w, r, e.ErrInvalidRequest(global_cats_sql.Error))
		return
	}

	cats, err := util.ScanAndSplitGlobalCats(global_cats_sql)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	var counts *[]model.CatCount
	if has_period {
		counts, err = util.GetCatCountsDuringPeriod(cats, period_params)
	} else {
		counts, err = util.GetCatCounts(cats)
	}

	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}
	util.RenderCategoryCounts(counts, w, r)
}

func AddTag(w http.ResponseWriter, r *http.Request) {
	tag_data := &model.NewTagRequest{}
	if err := render.Bind(r, tag_data); err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}
	
	link_exists, err := util.LinkExists(tag_data.LinkID)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	} else if !link_exists {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoLinkWithID))
		return
	}

	req_login_name := r.Context().Value(m.LoginNameKey).(string)
	duplicate, err := util.UserHasTaggedLink(req_login_name, tag_data.LinkID)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	} else if duplicate {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrDuplicateTag))
		return
	}
	
	tag_data.Categories = strings.ToLower(tag_data.Categories)
	res, err := db.Client.Exec(
		"INSERT INTO Tags VALUES(?,?,?,?,?);", 
		nil, 
		tag_data.LinkID, 
		tag_data.Categories, 
		req_login_name, 
		tag_data.LastUpdated,
	)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	if err = util.CalculateAndSetGlobalCats(tag_data.LinkID); err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	if err := util.AssignNewTagIDToRequest(res, tag_data); err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, tag_data)
}

// EDIT TAG
func EditTag(w http.ResponseWriter, r *http.Request) {
	edit_tag_data := &model.EditTagRequest{}
	if err := render.Bind(r, edit_tag_data); err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	req_login_name := r.Context().Value(m.LoginNameKey).(string)
	owns_tag, err := util.UserSubmittedTagWithID(req_login_name, edit_tag_data.ID)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoTagWithID))
		return
	} else if !owns_tag {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrDoesntOwnTag))
		return
	}

	edit_tag_data.Categories = util.AlphabetizeCats(edit_tag_data.Categories)

	_, err = db.Client.Exec(
		`UPDATE Tags 
		SET categories = ?, last_updated = ? 
		WHERE id = ?;`, 
		edit_tag_data.Categories, 
		edit_tag_data.LastUpdated, 
		edit_tag_data.ID,
	)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	link_id, err := util.GetLinkIDFromTagID(edit_tag_data.ID)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	} else if err = util.CalculateAndSetGlobalCats(link_id); err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, edit_tag_data)

}