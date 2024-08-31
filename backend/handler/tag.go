package handler

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"github.com/julianlk522/fitm/db"
	e "github.com/julianlk522/fitm/error"
	util "github.com/julianlk522/fitm/handler/util"
	m "github.com/julianlk522/fitm/middleware"
	"github.com/julianlk522/fitm/model"
	modelutil "github.com/julianlk522/fitm/model/util"
	"github.com/julianlk522/fitm/query"
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
	} else if !link_exists {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoLinkWithID))
		return
	}


	// refresh global cats before querying
	util.CalculateAndSetGlobalCats(link_id)

	req_user_id := r.Context().Value(m.UserIDKey).(string)
	link_sql := query.NewTagPageLink(link_id)
	if req_user_id != "" {
		link_sql = link_sql.AsSignedInUser(req_user_id)
	}
	if link_sql.Error != nil {
		render.Render(w, r, e.ErrInvalidRequest(link_sql.Error))
		return
	}

	req_login_name := r.Context().Value(m.LoginNameKey).(string)
	user_tag, err := util.GetUserTagForLink(req_login_name, link_id)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	tag_rankings_sql := query.NewTagRankings(link_id).Public()
	if tag_rankings_sql.Error != nil {
		render.Render(w, r, e.ErrInvalidRequest(tag_rankings_sql.Error))
		return
	}

	tag_rankings, err := util.ScanPublicTagRankings(tag_rankings_sql)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	if req_user_id != "" {
		link, err := util.ScanTagPageLink[model.LinkSignedIn](link_sql)
		if err != nil {
			render.Render(w, r, e.ErrInvalidRequest(err))
			return
		}

		render.JSON(w, r, model.TagPage[model.LinkSignedIn]{
			Link:        link,
			UserTag:     user_tag,
			TagRankings: tag_rankings,
		})

	} else {
		link, err := util.ScanTagPageLink[model.Link](link_sql)
		if err != nil {
			render.Render(w, r, e.ErrInvalidRequest(err))
			return
		}

		render.JSON(w, r, model.TagPage[model.Link]{
			Link:        link,
			UserTag:     user_tag,
			TagRankings: tag_rankings,
		})
	}
	
}

func GetTopGlobalCats(w http.ResponseWriter, r *http.Request) {
	global_cats_sql := query.NewTopGlobalCatCounts()

	// cats_params used to query subcats of cats
	cats_params := r.URL.Query().Get("cats")
	if cats_params != "" {
		global_cats_sql = global_cats_sql.SubcatsOfCats(cats_params)
	}

	period_params := r.URL.Query().Get("period")
	if period_params != "" {
		global_cats_sql = global_cats_sql.DuringPeriod(period_params)
	}

	if global_cats_sql.Error != nil {
		render.Render(w, r, e.ErrInvalidRequest(global_cats_sql.Error))
		return
	}

	counts, err := util.ScanGlobalCatCounts(global_cats_sql)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}
	util.RenderCatCounts(counts, w, r)
}

func GetTopContributors(w http.ResponseWriter, r *http.Request) {
	contributors_sql := query.NewContributors()

	cats_params := r.URL.Query().Get("cats")
	if cats_params != "" {
		cats := strings.Split(cats_params, ",")
		contributors_sql = contributors_sql.FromCats(cats)
	}
	

	period_params := r.URL.Query().Get("period")
	if period_params != "" {
		contributors_sql = contributors_sql.DuringPeriod(period_params)
	}

	if contributors_sql.Error != nil {
		render.Render(w, r, e.ErrInvalidRequest(contributors_sql.Error))
		return
	}

	contributors := util.ScanContributors(contributors_sql)
	util.RenderContributors(contributors, w, r)
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

	tag_data.Cats = util.AlphabetizeCats(tag_data.Cats)

	_, err = db.Client.Exec(
		"INSERT INTO Tags VALUES(?,?,?,?,?);",
		tag_data.ID,
		tag_data.LinkID,
		tag_data.Cats,
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

	// trim spaces
	edit_tag_data.Cats = modelutil.TrimExcessAndTrailingSpaces(edit_tag_data.Cats)

	// alphabetize
	edit_tag_data.Cats = util.AlphabetizeCats(edit_tag_data.Cats)

	_, err = db.Client.Exec(
		`UPDATE Tags 
		SET cats = ?, last_updated = ? 
		WHERE id = ?;`,
		edit_tag_data.Cats,
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
