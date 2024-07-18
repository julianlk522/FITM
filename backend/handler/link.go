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
	"oitm/model"
)

func GetTopLinks(w http.ResponseWriter, r *http.Request) {
	get_links_sql := query.NewGetTopLinks().Limit(LINKS_PAGE_LIMIT)
	if get_links_sql.Error != nil {
		render.Render(w, r, ErrInvalidRequest(get_links_sql.Error))
		return
	}

	links, err := _ScanLinks(get_links_sql, r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	_RenderLinks(links, w, r)
}

func GetTopLinksByPeriod(w http.ResponseWriter, r *http.Request) {
	period_params := chi.URLParam(r, "period")
	if period_params == "" {
		render.Render(w, r, ErrInvalidRequest(ErrNoPeriod))
		return
	}
	
	get_links_sql := query.NewGetTopLinks().DuringPeriod(period_params).Limit(LINKS_PAGE_LIMIT)
	if get_links_sql.Error != nil {
		render.Render(w, r, ErrInvalidRequest(get_links_sql.Error))
		return
	}

	links, err := _ScanLinks(get_links_sql, r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	_RenderLinks(links, w, r)
}

func GetTopLinksByCategories(w http.ResponseWriter, r *http.Request) {
	categories_params := chi.URLParam(r, "categories")
	if categories_params == "" {
		render.Render(w, r, ErrInvalidRequest(ErrNoCategories))
		return
	}

	link_ids, err := _GetIDsOfLinksHavingCategories(categories_params)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if len(link_ids) == 0 {
		_RenderZeroLinks(w, r)
		return
	}

	get_links_sql := query.NewGetTopLinks().FromLinkIDs(link_ids).Limit(LINKS_PAGE_LIMIT)
	if get_links_sql.Error != nil {
		render.Render(w, r, ErrInvalidRequest(get_links_sql.Error))
		return
	}

	links, err := _ScanLinks(get_links_sql, r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	_RenderLinks(links, w, r)
}

func GetTopLinksByPeriodAndCategories(w http.ResponseWriter, r *http.Request) {
	period_params, categories_params := chi.URLParam(r, "period"), chi.URLParam(r, "categories")
	if period_params == "" {
		render.Render(w, r, ErrInvalidRequest(ErrNoPeriod))
		return
	} else if categories_params == "" {
		render.Render(w, r, ErrInvalidRequest(ErrNoCategories))
		return
	}

	link_ids, err := _GetIDsOfLinksHavingCategories(categories_params)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if len(link_ids) == 0 {
		_RenderZeroLinks(w, r)
		return
	}

	get_links_sql := query.NewGetTopLinks().FromLinkIDs(link_ids).DuringPeriod(period_params).Limit(LINKS_PAGE_LIMIT)
	if get_links_sql.Error != nil {
		render.Render(w, r, ErrInvalidRequest(get_links_sql.Error))
		return
	}
	
	links, err := _ScanLinks(get_links_sql, r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	_RenderLinks(links, w, r)
}

func _ScanLinks(get_links_sql *query.GetTopLinks, r *http.Request) (*[]model.Link, error) {
	req_user_id, _, err := GetJWTClaims(r)
	if err != nil {
		return nil, err
	}

	rows, err := DBClient.Query(get_links_sql.Text)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := []model.Link{}

	// Auth: Add IsLiked / IsCopied / IsTagged to links
	if req_user_id != "" {
		for rows.Next() {

			// Note: I found it impossible to reduce this repeated code without upsetting the compiler... may come back after learning more
			i := model.LinkSignedIn{}
			if err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.Categories, &i.Summary, &i.SummaryCount, &i.LikeCount, &i.ImgURL); err != nil {
				return nil, err
			}
	
			// Add IsLiked / IsCopied / IsTagged
			var l sql.NullInt32
			var t sql.NullInt32
			var c sql.NullInt32
	
			
			err := DBClient.QueryRow(fmt.Sprintf(`SELECT
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
			) as is_copied;`, i.ID, req_user_id)).Scan(&l,&t, &c)
			if err != nil {
				return nil, err
			}
	
			i.IsLiked = l.Int32 > 0
			i.IsTagged = t.Int32 > 0
			i.IsCopied = c.Int32 > 0
	
			links = append(links, i)
		}

	// No auth
	} else {
		for rows.Next() {
			i := model.LinkSignedOut{}
			if err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.Categories, &i.Summary, &i.SummaryCount, &i.LikeCount, &i.ImgURL); err != nil {
				log.Fatal(err)
			}
	
			links = append(links, i)
		}
	}

	return &links, err
}

func _RenderLinks(links *[]model.Link, w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, &links)
	render.Status(r, http.StatusOK)
}

func _RenderZeroLinks(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, []model.Link{})
	render.Status(r, http.StatusOK)
}

func _GetIDsOfLinksHavingCategories(categories_str string) (link_ids []string, err error) {
	get_link_ids_sql := query.NewGetLinkIDs(categories_str)
	if get_link_ids_sql.Error != nil {
		err = get_link_ids_sql.Error
	}

	rows, err := DBClient.Query(get_link_ids_sql.Text)
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

func GetTopCategoryContributors(w http.ResponseWriter, r *http.Request) {
	categories_params := chi.URLParam(r, "categories")
	if categories_params == "" {
		render.Render(w, r, ErrInvalidRequest(ErrNoCategories))
		return
	}
	categories := strings.Split(categories_params, ",")

	get_contributors_sql := query.NewGetCategoryContributors(categories).Limit(CATEGORY_CONTRIBUTORS_LIMIT)
	if get_contributors_sql.Error != nil {
		render.Render(w, r, ErrInvalidRequest(get_contributors_sql.Error))
		return
	}
	
	contributors := _ScanCategoryContributors(get_contributors_sql, categories_params)
	_RenderCategoryContributors(contributors, w, r)
}

func GetTopCategoryContributorsByPeriod(w http.ResponseWriter, r *http.Request) {
	period_params, categories_params := chi.URLParam(r, "period"), chi.URLParam(r, "categories")
	if period_params == "" {
		render.Render(w, r, ErrInvalidRequest(ErrNoPeriod))
		return
	} else if categories_params == "" {
		render.Render(w, r, ErrInvalidRequest(ErrNoCategories))
		return
	}
	
	categories := strings.Split(categories_params, ",")
	get_contributors_sql := query.NewGetCategoryContributors(categories).DuringPeriod(period_params).Limit(CATEGORY_CONTRIBUTORS_LIMIT)
	if get_contributors_sql.Error != nil {
		render.Render(w, r, ErrInvalidRequest(get_contributors_sql.Error))
		return
	}

	contributors := _ScanCategoryContributors(get_contributors_sql, categories_params)
	_RenderCategoryContributors(contributors, w, r)
}

func _ScanCategoryContributors(get_contributors_sql *query.GetCategoryContributors, categories_str string) *[]model.CategoryContributor {
	rows, err := DBClient.Query(get_contributors_sql.Text)
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

func GetSubcategories(w http.ResponseWriter, r *http.Request) {
	categories_params := chi.URLParam(r, "categories")
	if categories_params == "" {
		render.Render(w, r, ErrInvalidRequest(ErrNoCategories))
		return
	}

	// TODO: replace with middleware that converts all URLs to lowercase
	// maybe encode uppercase chars another way?
	// TODO: figure out how other sites do that
	categories_params = strings.ToLower(categories_params)
	categories := strings.Split(categories_params, ",")
	
	get_subcats_sql := query.NewGetSubcategories(categories)
	if get_subcats_sql.Error != nil {
		render.Render(w, r, ErrInvalidRequest(get_subcats_sql.Error))
		return
	}

	subcats := _ScanSubcategories(get_subcats_sql, categories)
	if len(subcats) == 0 {
		_RenderZeroSubcategories(w, r)
		return
	}
	_RenderSubcategories(subcats, categories, w, r)
}

func GetSubcategoriesByPeriod(w http.ResponseWriter, r *http.Request) {
	period_params, categories_params := chi.URLParam(r, "period"), chi.URLParam(r, "categories")
	if period_params == "" {
		render.Render(w, r, ErrInvalidRequest(ErrNoPeriod))
		return
	} else if categories_params == "" {
		render.Render(w, r, ErrInvalidRequest(ErrNoCategories))
		return
	}
	categories_params = strings.ToLower(categories_params)
	categories := strings.Split(categories_params, ",")

	get_subcats_sql := query.NewGetSubcategories(categories).DuringPeriod(period_params)
	if get_subcats_sql.Error != nil {
		render.Render(w, r, ErrInvalidRequest(get_subcats_sql.Error))
		return
	}

	subcats := _ScanSubcategories(get_subcats_sql, categories)
	if len(subcats) == 0 {
		_RenderZeroSubcategories(w, r)
		return
	}
	_RenderSubcategories(subcats, categories, w, r)
}

func _ScanSubcategories(get_subcats_sql *query.GetSubcategories, search_categories []string) []string {
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
		render.Render(w, r, ErrInvalidRequest(err))
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
		get_link_count_sql := query.NewGetLinkCount(all_cats)
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

	if len(*cats_with_counts) > CATEGORY_COUNT_LIMIT {
		*cats_with_counts = (*cats_with_counts)[:CATEGORY_COUNT_LIMIT]
	}
}

func AddLink(w http.ResponseWriter, r *http.Request) {
	link_data := &model.NewLinkRequest{}
	if err := render.Bind(r, link_data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if strings.Count(link_data.NewLink.Categories, ",") > NEW_TAG_CATEGORY_LIMIT {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("too many tag categories (%d max)", NEW_TAG_CATEGORY_LIMIT)))
		return
	}

    resp, err := _ResolveURL(link_data)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if _URLAlreadySaved(link_data.URL) {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("duplicate URL: %s", link_data.URL)))
		return
	}

	req_user_id, req_login_name, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	link_data.SubmittedBy = req_login_name

	_AssignMetadata(link_data, req_user_id, resp)
	_AssignSortedCategories(link_data, link_data.NewLink.Categories)

	res, err := DBClient.Exec("INSERT INTO Links VALUES(?,?,?,?,?,?,?);", nil, link_data.URL, req_login_name, link_data.SubmitDate, link_data.Categories, link_data.Summary, link_data.ImgURL)
	if err != nil {
		log.Fatal(err)
	}

	if err := _AssignID(link_data, res); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if link_data.Summary != "" {
		_, err = DBClient.Exec("INSERT INTO Summaries VALUES(?,?,?,?,?);", nil, link_data.Summary, link_data.ID, link_data.SummaryAuthor, link_data.SubmitDate)
		if err != nil {
			log.Fatal(err)
		}

		link_data.SummaryCount = 1
	}

	_, err = DBClient.Exec("INSERT INTO Tags VALUES(?,?,?,?,?);", nil, link_data.ID, link_data.Categories, req_login_name, link_data.SubmitDate)
	if err != nil {
		log.Fatal(err)
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, link_data)
}

func _ResolveURL(link_data *model.NewLinkRequest) (*http.Response, error) {	
	has_protocol_regex, err := regexp.Compile(`^(http(s?)\:\/\/)`)
	if err != nil {
		return nil, err
	}
	
	var resp *http.Response
	var ErrRedirect error = errors.New("invalid link destination: redirect detected")

	// Protocol specified: check as-is
	if has_protocol_regex.MatchString(link_data.NewLink.URL) {
		resp, err = http.Get(link_data.NewLink.URL)
		if _IsRedirect(resp.StatusCode) {
			return nil, ErrRedirect
		} else if err != nil || resp.StatusCode == 404 {
			return nil, _InvalidURLError(link_data.URL)
		}
		
	// Protocol not specified: try https then http
	} else {
	
		// https
		link_data.URL = "https://" + link_data.NewLink.URL
		resp, err = http.Get(link_data.URL)
		if err != nil {

			// http
			link_data.URL = "http://" + link_data.NewLink.URL
			resp, err = http.Get(link_data.URL)
			if _IsRedirect(resp.StatusCode) {
				return nil, ErrRedirect
			} else if err != nil {
				return nil, _InvalidURLError(link_data.URL)
			}

		} else if _IsRedirect(resp.StatusCode) {
			return nil, ErrRedirect
		}
	}
	
	// Valid URL: save after any redirects e.g., to wwww.
	link_data.URL = resp.Request.URL.String()

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

func _AssignMetadata(link_data *model.NewLinkRequest, req_user_id string, resp *http.Response) {
	defer resp.Body.Close()

	meta := MetaFromHTMLTokens(resp.Body)

	if link_data.Summary != "" {
		link_data.SummaryAuthor = req_user_id

	// Auto Summary
	} else {
		switch {
			case meta.OGDescription != "":
				link_data.Summary = meta.OGDescription
			case meta.Description != "":
				link_data.Summary = meta.Description
			case meta.OGTitle != "":
				link_data.Summary = meta.OGTitle
			case meta.Title != "":
				link_data.Summary = meta.Title
			case meta.OGSiteName != "":
				link_data.Summary = meta.OGSiteName
		}

		if link_data.Summary != "" {

			// 15 is Auto Summary's user_id
			// TODO: update with final
			link_data.SummaryAuthor = "15"
		}
	}

	if meta.OGImage != "" {
		resp, err := http.Get(meta.OGImage)
		if err == nil && resp.StatusCode != 404 && !_IsRedirect(resp.StatusCode) {
			link_data.ImgURL = meta.OGImage
		}
	}
}

func _AssignSortedCategories(link *model.NewLinkRequest, categories_str string) {
	split_categories := strings.Split(categories_str, ",")
	slices.Sort(split_categories)

	link.Categories = strings.Join(split_categories, ",")
}

func _AssignID(link *model.NewLinkRequest, res sql.Result) error {
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	link.ID = id
	return nil
}

// LIKE LINK
func LikeLink(w http.ResponseWriter, r *http.Request) {
	link_id := chi.URLParam(r, "link_id")
	if link_id == "" {
		render.Render(w, r, ErrInvalidRequest(ErrInvalidLinkID))
		return
	}

	req_user_id, req_login_name, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if _UserSubmittedLink(req_login_name, link_id) {
		render.Render(w, r, ErrInvalidRequest(errors.New("cannot like your own link")))
		return
	}

	if _UserHasLikedLink(req_user_id, link_id) {
		render.Render(w, r, ErrInvalidRequest(errors.New("already liked")))
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
		render.Render(w, r, ErrInvalidRequest(ErrInvalidLinkID))
		return
	}

	req_user_id, _, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if !_UserHasLikedLink(req_user_id, link_id) {
		render.Render(w, r, ErrInvalidRequest(errors.New("link like not found")))
		return
	}

	_, err = DBClient.Exec("DELETE FROM 'Link Likes' WHERE user_id = ? AND link_id = ?;", req_user_id, link_id)
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
		render.Render(w, r, ErrInvalidRequest(ErrNoLinkID))
		return
	}

	req_user_id, _, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	already_copied := _UserHasCopiedLink(req_user_id, link_id)
	if already_copied {
		render.Render(w, r, ErrInvalidRequest(errors.New("link already copied to treasure map")))
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
		render.Render(w, r, ErrInvalidRequest(ErrNoLinkID))
		return
	}

	req_user_id, _, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	already_copied := _UserHasCopiedLink(req_user_id, link_id)
	if !already_copied {
		render.Render(w, r, ErrInvalidRequest(errors.New("link copy does not exist")))
		return
	}

	// Delete
	_, err = DBClient.Exec("DELETE FROM 'Link Copies' WHERE user_id = ? AND link_id = ?;", req_user_id, link_id)
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