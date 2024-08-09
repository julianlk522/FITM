package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"golang.org/x/exp/slices"

	query "oitm/db/query"
	e "oitm/error"
	util "oitm/handler/util"
	m "oitm/middleware"
	"oitm/model"
)

func GetLinks(w http.ResponseWriter, r *http.Request) {
	links_sql := query.NewTopLinks()

	cats_params := r.URL.Query().Get("cats")
	if cats_params != "" {
		link_ids, err := _GetIDsOfLinksHavingCategories(cats_params)
		if err != nil {
			render.Render(w, r, e.ErrInvalidRequest(err))
			return
		} else if len(link_ids) == 0 {
			_RenderZeroLinks(w, r)
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
		links, err := _ScanLinks[model.LinkSignedIn](links_sql, req_user_id)
		if err != nil {
			render.Render(w, r, e.ErrInvalidRequest(err))
			return
		}
		_RenderPaginatedLinks(links, page, w, r)
	} else {
		links, err := _ScanLinks[model.Link](links_sql, req_user_id)
		if err != nil {
			render.Render(w, r, e.ErrInvalidRequest(err))
			return
		}
		_RenderPaginatedLinks(links, page, w, r)
	}

}

func _GetIDsOfLinksHavingCategories(categories_str string) (link_ids []string, err error) {
	link_ids_sql := query.NewLinkIDs(categories_str)
	if link_ids_sql.Error != nil {
		err = link_ids_sql.Error
	}

	rows, err := DBClient.Query(link_ids_sql.Text)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var lid string
		if err := rows.Scan(&lid); err != nil {
			log.Fatal(err)
		}

		link_ids = append(link_ids, lid)
	}

	return link_ids, err
}

func _RenderZeroLinks(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, &model.PaginatedLinks[model.Link]{NextPage: -1})
	render.Status(r, http.StatusOK)
}

func _ScanLinks[T model.LinkSignedIn | model.Link](get_links_sql *query.TopLinks, req_user_id string) (*[]T, error) {
	var links interface{}

	rows, err := DBClient.Query(get_links_sql.Text)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	switch any(new(T)).(type) {
	case *model.LinkSignedIn:
		var signed_in_links = []model.LinkSignedIn{}
	
		for rows.Next() {
			i := model.LinkSignedIn{}
			err := rows.Scan(
				&i.ID, 
				&i.URL, 
				&i.SubmittedBy, 
				&i.SubmitDate, 
				&i.Categories, 
				&i.Summary, 
				&i.SummaryCount, 
				&i.TagCount,
				&i.LikeCount, 
				&i.ImgURL,
			)
			if err != nil {
				return nil, err
			}

			// Add IsLiked / IsCopied / IsTagged
			var l sql.NullInt32
			var t sql.NullInt32
			var c sql.NullInt32
	
			
			err = DBClient.QueryRow(
				fmt.Sprintf(
					`SELECT
					(
						SELECT count(*) FROM 'Link Likes'
						WHERE link_id = '%[1]d' AND user_id = '%[2]s'
					) as is_liked,
					(
						SELECT count(*) FROM Tags
						JOIN Users
						ON Users.login_name = Tags.submitted_by
						WHERE link_id = '%[1]d' AND Users.id = '%[2]s'
					) AS is_tagged,
					(
						SELECT count(*) FROM 'Link Copies'
						WHERE link_id = '%[1]d' AND user_id = '%[2]s'
					) as is_copied;`, 
					i.ID, 
					req_user_id,
				),
			).Scan(&l,&t, &c)
			if err != nil {
				return nil, err
			}
	
			i.IsLiked = l.Int32 > 0
			i.IsTagged = t.Int32 > 0
			i.IsCopied = c.Int32 > 0

			signed_in_links = append(signed_in_links, i)
		}
	
		links = &signed_in_links
			
	case *model.Link:
		var signed_out_links = []model.Link{}
	
		for rows.Next() {
			i := model.Link{}
			err := rows.Scan(
				&i.ID, 
				&i.URL, 
				&i.SubmittedBy, 
				&i.SubmitDate, 
				&i.Categories, 
				&i.Summary, 
				&i.SummaryCount, 
				&i.TagCount,
				&i.LikeCount, 
				&i.ImgURL,
			)
			if err != nil {
				return nil, err
			}
			signed_out_links = append(signed_out_links, i)
		}
	
		links = &signed_out_links
	}
	
	return links.(*[]T), nil
}

func _RenderPaginatedLinks[T model.LinkSignedIn | model.Link](links *[]T, page int, w http.ResponseWriter, r *http.Request) {
	if len(*links) == 0 {
		_RenderZeroLinks(w, r)
	} else if len(*links) == LINKS_PAGE_LIMIT + 1 {
		sliced := (*links)[:LINKS_PAGE_LIMIT]
		render.JSON(w, r, &model.PaginatedLinks[T]{
			Links: &sliced, 
			NextPage: page + 1,
		})
	} else {
		render.JSON(w, r, &model.PaginatedLinks[T]{
			Links: links, 
			NextPage: -1,
		})
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

	contributors_sql = contributors_sql.Limit(CATEGORY_CONTRIBUTORS_LIMIT)

	if contributors_sql.Error != nil {
		render.Render(w, r, e.ErrInvalidRequest(contributors_sql.Error))
		return
	}
	
	contributors := _ScanCategoryContributors(contributors_sql, cats_params)
	_RenderCategoryContributors(contributors, w, r)
}

func _ScanCategoryContributors(contributors_sql *query.CatsContributors, categories_str string) *[]model.CategoryContributor {
	rows, err := DBClient.Query(contributors_sql.Text)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	contributors := []model.CategoryContributor{}
	for rows.Next() {
		contributor := model.CategoryContributor{Categories: categories_str}
		err := rows.Scan(&contributor.LinksSubmitted, &contributor.LoginName)
		if err != nil {
			log.Fatal(err)
		}
		contributors = append(contributors, contributor)
	}

	return &contributors
}

func _RenderCategoryContributors(contributors *[]model.CategoryContributor, w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, contributors)
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

	subcats := _ScanSubcategories(subcats_sql, categories)
	if len(subcats) == 0 {
		_RenderZeroSubcategories(w, r)
		return
	}
	_RenderSubcategories(subcats, categories, w, r)
}

func _ScanSubcategories(get_subcats_sql *query.Subcats, search_categories []string) []string {
	rows, err := DBClient.Query(get_subcats_sql.Text)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var subcats []string
	for rows.Next() {
		var row_cats string
		if err := rows.Scan(&row_cats); err != nil {
			log.Fatal(err)
		}

		cats := strings.Split(row_cats, ",")
		for i := 0; i < len(cats); i++ {
			cat_lc := strings.ToLower(cats[i])

			if !slices.Contains(search_categories, cat_lc) && !slices.Contains(subcats, cat_lc) {
				subcats = append(subcats, cat_lc)
			}
		}
	}
	
	return subcats
}

func _RenderZeroSubcategories(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, []model.CategoryCount{})
	render.Status(r, http.StatusOK)
}

func _RenderSubcategories(subcats []string, categories []string, w http.ResponseWriter, r *http.Request) {
	with_counts, err := _GetSubcategoryCounts(subcats, categories)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	_SortAndLimitCategoryCounts(with_counts)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, with_counts)
}

func _GetSubcategoryCounts(subcats []string, categories []string) (*[]model.CategoryCount, error) {
	subcats_with_counts := make([]model.CategoryCount, len(subcats))
	for i := 0; i < len(subcats); i++ {
		subcats_with_counts[i].Category = subcats[i]
		
		all_cats := append(categories, subcats[i])
		get_link_count_sql := query.NewCatCount(all_cats)
		if get_link_count_sql.Error != nil {
			return nil, get_link_count_sql.Error
		}

		if err := DBClient.QueryRow(get_link_count_sql.Text).Scan(&subcats_with_counts[i].Count); err != nil {
			return nil, err
		}
	}

	return &subcats_with_counts, nil
}

func _SortAndLimitCategoryCounts(cats_with_counts *[]model.CategoryCount) {
	slices.SortFunc(*cats_with_counts, model.SortCategories)

	if len(*cats_with_counts) > CATEGORY_PAGE_LIMIT {
		*cats_with_counts = (*cats_with_counts)[:CATEGORY_PAGE_LIMIT]
	}
}

func AddLink(w http.ResponseWriter, r *http.Request) {
	request := &model.NewLinkRequest{}
	if err := render.Bind(r, request); err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	// Check URL valid and unique
    resp, err := _ResolveAndAssignURL(request.NewLink.URL, request)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	if _URLAlreadySaved(request.URL) {
		render.Render(w, r, e.ErrInvalidRequest(fmt.Errorf("duplicate URL: %s", request.URL)))
		return
	}

	req_login_name := r.Context().Value(m.LoginNameKey).(string)
	request.SubmittedBy = req_login_name
	
	_AssignMetadata(resp, request)

	unsorted_cats := request.NewLink.Categories
	_AssignSortedCategories(unsorted_cats, request)
	
	// Insert link
	res, err := DBClient.Exec(
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
	
	if err := _AssignNewLinkID(res, request); err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	// Insert Summary(ies)
	// (might have user-submitted, Auto Summary, or both)
	if request.AutoSummary != "" {
		// Note: UserID 15 is AutoSummary
		// TODO: add constant, replace magic 15
		_, err = DBClient.Exec(
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
		_, err = DBClient.Exec(
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
	_, err = DBClient.Exec(
		"INSERT INTO Tags VALUES(?,?,?,?,?);", 
		nil, 
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

func _ResolveAndAssignURL(url string, request *model.NewLinkRequest) (*http.Response, error) {	
	has_protocol_regex, err := regexp.Compile(`^(http(s?)\:\/\/)`)
	if err != nil {
		return nil, err
	}
	
	var resp *http.Response
	var ErrRedirect error = errors.New("invalid link destination: redirect detected")

	// Protocol specified: check as-is
	if has_protocol_regex.MatchString(url) {
		resp, err = http.Get(url)
		if _IsRedirect(resp.StatusCode) {
			return nil, ErrRedirect
		} else if err != nil || resp.StatusCode == 404 {
			return nil, _InvalidURLError(url)
		}
		
	// Protocol not specified: try https then http
	} else {

		// https
		modified_url := "https://" + url
		resp, err = http.Get(modified_url)
		if err != nil {

			// http
			modified_url = "http://" + url
			resp, err = http.Get(modified_url)
			if _IsRedirect(resp.StatusCode) {
				return nil, ErrRedirect
			} else if err != nil {
				return nil, _InvalidURLError(modified_url)
			}

		} else if _IsRedirect(resp.StatusCode) {
			return nil, ErrRedirect
		}
	}
	
	// save updated URL after any redirects e.g., to wwww.
	request.URL = resp.Request.URL.String()

	return resp, nil
}

func _IsRedirect(status_code int) bool {
	return status_code > 299 && status_code < 400
}

func _InvalidURLError(url string) error {
	return fmt.Errorf("invalid URL: %s", url)
}

func _URLAlreadySaved(url string) bool {
	var u sql.NullString

	err := DBClient.QueryRow("SELECT url FROM Links WHERE url = ?", url).Scan(&u)
	return err == nil && u.Valid
}

func _AssignMetadata(resp *http.Response, link_data *model.NewLinkRequest) {
	defer resp.Body.Close()

	meta := util.MetaFromHTMLTokens(resp.Body)

	switch {
		case meta.OGDescription != "":
			link_data.AutoSummary = meta.OGDescription
		case meta.Description != "":
			link_data.AutoSummary = meta.Description
		case meta.OGTitle != "":
			link_data.AutoSummary = meta.OGTitle
		case meta.Title != "":
			link_data.AutoSummary = meta.Title
		case meta.OGSiteName != "":
			link_data.AutoSummary = meta.OGSiteName
	}

	if meta.OGImage != "" {
		resp, err := http.Get(meta.OGImage)
		if err == nil && resp.StatusCode != 404 && !_IsRedirect(resp.StatusCode) {
			link_data.ImgURL = meta.OGImage
		}
	}
}

func _AssignSortedCategories(unsorted_cats string, link *model.NewLinkRequest) {
	split_categories := strings.Split(unsorted_cats, ",")
	slices.Sort(split_categories)

	link.Categories = strings.Join(split_categories, ",")
}

func _AssignNewLinkID(res sql.Result, request *model.NewLinkRequest) error {
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	request.ID = id
	return nil
}

// LIKE LINK
func LikeLink(w http.ResponseWriter, r *http.Request) {
	link_id := chi.URLParam(r, "link_id")
	if link_id == "" {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoLinkID))
		return
	}

	req_login_name := r.Context().Value(m.LoginNameKey).(string)
	if _UserSubmittedLink(req_login_name, link_id) {
		render.Render(w, r, e.ErrInvalidRequest(errors.New("cannot like your own link")))
		return
	}

	req_user_id := r.Context().Value(m.UserIDKey).(string)
	if _UserHasLikedLink(req_user_id, link_id) {
		render.Render(w, r, e.ErrInvalidRequest(errors.New("already liked")))
		return
	}

	res, err := DBClient.Exec("INSERT INTO 'Link Likes' VALUES(?,?,?);", nil, req_user_id, link_id)
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

// UN-LIKE LINK
func UnlikeLink(w http.ResponseWriter, r *http.Request) {
	link_id := chi.URLParam(r, "link_id")
	if link_id == "" {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoLinkID))
		return
	}

	req_user_id := r.Context().Value(m.UserIDKey).(string)
	if !_UserHasLikedLink(req_user_id, link_id) {
		render.Render(w, r, e.ErrInvalidRequest(errors.New("link like not found")))
		return
	}

	_, err := DBClient.Exec("DELETE FROM 'Link Likes' WHERE user_id = ? AND link_id = ?;", req_user_id, link_id)
	if err != nil {
		log.Fatal(err)
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{"message": "deleted"})
}

func _UserSubmittedLink(login_name string, link_id string) bool {
	var link_submitted_by sql.NullString
	err := DBClient.QueryRow("SELECT submitted_by FROM Links WHERE id = ?;", link_id).Scan(&link_submitted_by)

	if err != nil {
		return false
	}

	return link_submitted_by.String == login_name
}

func _UserHasLikedLink(user_id string, link_id string) bool {
	var l sql.NullString
	err := DBClient.QueryRow("SELECT id FROM 'Link Likes' WHERE user_id = ? AND link_id = ?;",user_id, link_id).Scan(&l)

	return err == nil && l.Valid
}

// COPY LINK TO USER'S TREASURE MAP
func CopyLink(w http.ResponseWriter, r *http.Request) {
	link_id := chi.URLParam(r, "link_id")
	if link_id == "" {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoLinkID))
		return
	}

	req_login_name := r.Context().Value(m.LoginNameKey).(string)
	owns_link := _UserSubmittedLink(req_login_name, link_id)
	if owns_link {
		render.Render(w, r, e.ErrInvalidRequest(errors.New("cannot copy your own link to your treasure map")))
		return
	}

	req_user_id := r.Context().Value(m.UserIDKey).(string)
	already_copied := _UserHasCopiedLink(req_user_id, link_id)
	if already_copied {
		render.Render(w, r, e.ErrInvalidRequest(errors.New("link already copied to treasure map")))
		return
	}

	res, err := DBClient.Exec("INSERT INTO 'Link Copies' VALUES(?,?,?);", nil, link_id, req_user_id)
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

// UN-COPY LINK
func UncopyLink(w http.ResponseWriter, r *http.Request) {
	link_id := chi.URLParam(r, "link_id")
	if link_id == "" {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoLinkID))
		return
	}

	req_user_id := r.Context().Value(m.UserIDKey).(string)
	already_copied := _UserHasCopiedLink(req_user_id, link_id)
	if !already_copied {
		render.Render(w, r, e.ErrInvalidRequest(errors.New("link copy does not exist")))
		return
	}

	// Delete
	_, err := DBClient.Exec("DELETE FROM 'Link Copies' WHERE user_id = ? AND link_id = ?;", req_user_id, link_id)
	if err != nil {
		log.Fatal(err)
	}

	return_json := map[string]string{
		"message": "deleted",
	}

	render.JSON(w, r, return_json)
	render.Status(r, http.StatusNoContent)
}

func _UserHasCopiedLink(user_id string, link_id string) bool {
	var l sql.NullString
	err := DBClient.QueryRow("SELECT id FROM 'Link Copies' WHERE user_id = ? AND link_id = ?;", user_id, link_id).Scan(&l)

	return err == nil && l.Valid
}