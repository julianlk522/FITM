package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"regexp"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"golang.org/x/exp/slices"

	"oitm/model"
)

// GET OVERALL MOST-LIKED LINKS
// (top 20 for now)
func GetTopLinks(w http.ResponseWriter, r *http.Request) {
	db ,err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
	const LIMIT string = "20"
	get_links_sql := fmt.Sprintf(`SELECT links_id as link_id, url, link_author as submitted_by, sd, categories, summary, coalesce(count(Summaries.id),0) as summary_count, like_count, img_url
	FROM 
		(
		SELECT Links.id as links_id, url, submitted_by as link_author, Links.submit_date as sd, coalesce(global_cats,"") as categories, coalesce(global_summary,"") as summary, coalesce(like_count,0) as like_count, coalesce(img_url,"") as img_url 
		FROM LINKS 
		LEFT JOIN 
			(
			SELECT link_id as likes_link_id, count(*) as like_count 
			FROM 'Link Likes'
			GROUP BY likes_link_id
			) 
		ON Links.id = likes_link_id
		) 
	LEFT JOIN Summaries 
	ON Summaries.link_id = links_id 
	GROUP BY links_id 
	ORDER BY like_count DESC, summary_count DESC, link_id DESC LIMIT %s;`, LIMIT)
	rows, err := db.Query(get_links_sql)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
			
	// Check auth token
	req_user_id, _, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Scan links
	// User signed in
	if req_user_id != "" {
		links := ScanLinksSignedIn(db, rows, req_user_id)
		render.JSON(w, r, &links)

	// User signed out: IsLiked / IsCopied / IsTagged not included		
	} else {
		links := ScanLinksSignedOut(db, rows)
		render.JSON(w, r, &links)
	}

	render.Status(r, http.StatusOK)
}

// GET MOST-LIKED LINKS DURING PERIOD
// (day, week, month)
// (top 20 for now)
func GetTopLinksByPeriod(w http.ResponseWriter, r *http.Request) {
	period_params := chi.URLParam(r, "period")
	if period_params == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("no period provided")))
		return
	}
	get_links_sql := `SELECT links_id as link_id, url, link_author as subitted_by, sd, categories, summary, coalesce(count(Summaries.id),0) as summary_count, like_count, img_url
	FROM
		(
		SELECT Links.id as links_id, url, submitted_by as link_author, submit_date as sd, coalesce(global_cats,"") as categories, coalesce(global_summary,"") as summary, coalesce(like_count,0) as like_count, coalesce(img_url,"") as img_url
		FROM LINKS
		LEFT JOIN 
			(
			SELECT link_id as likes_link_id, count(*) as like_count
			FROM 'Link Likes'
			GROUP BY likes_link_id
			)
		ON Links.id = likes_link_id`

	AppendPeriodClause(&get_links_sql, period_params)

	const LIMIT string = "20"
	get_links_sql += fmt.Sprintf(`) LEFT JOIN Summaries
	ON Summaries.link_id = links_id
	GROUP BY links_id ORDER BY like_count DESC, summary_count DESC, link_id DESC LIMIT %s;`, LIMIT)

	db ,err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query(get_links_sql)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	req_user_id, _, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	
	// Scan links
	// User signed in
	if req_user_id != "" {
		links := ScanLinksSignedIn(db, rows, req_user_id)
		render.JSON(w, r, &links)

	// User signed out: IsLiked / IsCopied / IsTagged not included		
	} else {
		links := ScanLinksSignedOut(db, rows)
		render.JSON(w, r, &links)
	}

	render.Status(r, http.StatusOK)
}

// GET MOST-LIKED LINKS WITH GIVEN CATEGORY(IES)
// (top 20 for now)
func GetTopLinksByCategories(w http.ResponseWriter, r *http.Request) {
	categories_params := chi.URLParam(r, "categories")
	if categories_params == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("no categories provided")))
		return
	}

	link_ids := GetIDsOfLinksHavingCategories(categories_params)
	if len(link_ids) == 0 {
		render.JSON(w, r, []model.LinkSignedOut{})
		render.Status(r, http.StatusOK)
		return
	}

	// get link data for each ID
	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// limit top 20 links for now
	const LIMIT string = "20"
	rows, err := db.Query(fmt.Sprintf(`SELECT links_id as link_id, url, link_author as submitted_by, sd, categories, summary, coalesce(count(Summaries.id),0) as summary_count, like_count, img_url
	FROM
		(
		SELECT Links.id as links_id, url, submitted_by as link_author, submit_date as sd, coalesce(global_cats,"") as categories, coalesce(global_summary,"") as summary, coalesce(like_count,0) as like_count, coalesce(img_url,"") as img_url
		FROM LINKS
		LEFT JOIN 
			(
			SELECT link_id as likes_link_id, count(*) as like_count
			FROM 'Link Likes' 
			GROUP BY likes_link_id
			)
		ON Links.id = likes_link_id 
		WHERE links_id IN (%s)
		)
	LEFT JOIN Summaries
	ON Summaries.link_id = links_id
	GROUP BY link_id
	ORDER BY like_count DESC, link_id ASC 
	LIMIT %s;`, strings.Join(link_ids, ","), LIMIT))
	if err != nil {
		log.Fatal(err)
	}

	req_user_id, _, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Scan links
	// User signed in
	if req_user_id != "" {
		links := ScanLinksSignedIn(db, rows, req_user_id)
		render.JSON(w, r, &links)

	// User signed out: IsLiked / IsCopied / IsTagged not included		
	} else {
		links := ScanLinksSignedOut(db, rows)
		render.JSON(w, r, &links)
	}

	render.Status(r, http.StatusOK)
}

func GetTopLinksByPeriodAndCategories(w http.ResponseWriter, r *http.Request) {
	period_params, categories_params := chi.URLParam(r, "period"), chi.URLParam(r, "categories")
	if period_params == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("no period provided")))
		return
	} else if categories_params == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("no categories provided")))
		return
	}

	link_ids := GetIDsOfLinksHavingCategories(categories_params)
	if len(link_ids) == 0 {
		render.JSON(w, r, []model.LinkSignedOut{})
		render.Status(r, http.StatusOK)
		return
	}

	get_links_sql := `SELECT links_id as link_id, url, link_author as subitted_by, sd, categories, summary, coalesce(count(Summaries.id),0) as summary_count, like_count, img_url
	FROM
		(
		SELECT Links.id as links_id, url, submitted_by as link_author, submit_date as sd, coalesce(global_cats,"") as categories, coalesce(global_summary,"") as summary, coalesce(like_count,0) as like_count, coalesce(img_url,"") as img_url
		FROM LINKS
		LEFT JOIN 
			(
			SELECT link_id as likes_link_id, count(*) as like_count
			FROM 'Link Likes'
			GROUP BY likes_link_id
			)
		ON Links.id = likes_link_id`

	AppendPeriodClause(&get_links_sql, period_params)
	get_links_sql += fmt.Sprintf(` AND links_id IN (%s)`, strings.Join(link_ids, ","))

	const LIMIT string = "20"
	get_links_sql += fmt.Sprintf(`) LEFT JOIN Summaries
	ON Summaries.link_id = links_id
	GROUP BY links_id ORDER BY like_count DESC, summary_count DESC, link_id DESC LIMIT %s;`, LIMIT)

	db ,err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query(get_links_sql)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	req_user_id, _, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	
	// Scan links
	// User signed in
	if req_user_id != "" {
		links := ScanLinksSignedIn(db, rows, req_user_id)
		render.JSON(w, r, &links)

	// User signed out: IsLiked / IsCopied / IsTagged not included		
	} else {
		links := ScanLinksSignedOut(db, rows)
		render.JSON(w, r, &links)
	}

	render.Status(r, http.StatusOK)
}

func GetIDsOfLinksHavingCategories(categories string) []string {
	var get_links_sql string

	// multiple categories
	if strings.Contains(categories, ",") {
		categories := strings.Split(categories, ",")

		// get link IDs
		get_links_sql = fmt.Sprintf(`select id from Links where ',' || global_cats || ',' like '%%,%s,%%'`, categories[0])

		for i := 1; i < len(categories); i++ {
			get_links_sql += fmt.Sprintf(` AND ',' || global_cats || ',' like '%%,%s,%%'`, categories[i])
		}

	// single category
	} else {

		// get link IDs
		get_links_sql = fmt.Sprintf(`select id from Links where ',' || global_cats || ',' like '%%,%s,%%'`, categories)
	}
	get_links_sql += ` group by id`

	db ,err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query(get_links_sql)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	
	var link_ids []string
	for rows.Next() {
		var link_id string
		err := rows.Scan(&link_id)
		if err != nil {
			log.Fatal(err)
		}
		link_ids = append(link_ids, link_id)
	}

	return link_ids
}

// GET TOP CONTRIBUTORS FOR GIVEN CATEGORY(IES)
// (determined by number of links submitted having ALL given categories in global_cats)
func GetTopCategoryContributors(w http.ResponseWriter, r *http.Request) {

	// Limit 5
	const LIMIT string = "5"

	db ,err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// get categories
	categories_params := chi.URLParam(r, "categories")
	categories := strings.Split(categories_params, ",")
	get_links_sql := fmt.Sprintf(`select count(*), submitted_by from Links where ',' || global_cats || ',' like '%%,%s,%%'`, categories[0])

		for i := 1; i < len(categories); i++ {
			get_links_sql += fmt.Sprintf(` AND ',' || global_cats || ',' like '%%,%s,%%'`, categories[i])
		}

	get_links_sql += fmt.Sprintf(` GROUP BY submitted_by ORDER BY count(*) DESC LIMIT %s;`, LIMIT)

	rows, err := db.Query(get_links_sql)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	contributors := []model.CategoryContributor{}
	for rows.Next() {
		contributor := model.CategoryContributor{Categories: categories_params}
		err := rows.Scan(&contributor.LinksSubmitted, &contributor.LoginName)
		if err != nil {
			log.Fatal(err)
		}
		contributors = append(contributors, contributor)
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, contributors)
}

func GetTopCategoryContributorsByPeriod(w http.ResponseWriter, r *http.Request) {
	period_params, categories_params := chi.URLParam(r, "period"), chi.URLParam(r, "categories")
	if period_params == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("no period provided")))
		return
	} else if categories_params == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("no categories provided")))
		return
	}
	categories := strings.Split(categories_params, ",")

	get_links_sql := `SELECT count(*), submitted_by
		FROM Links`
	AppendPeriodClause(&get_links_sql, period_params)
	for _, cat := range categories {
		get_links_sql += fmt.Sprintf(` AND ',' || global_cats || ',' like '%%,%s,%%'`, cat)
	}

	// Limit 5
	const LIMIT string = "5"
	get_links_sql += fmt.Sprintf(` GROUP BY submitted_by ORDER BY count(*) DESC LIMIT %s;`, LIMIT)

	db ,err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query(get_links_sql)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	contributors := []model.CategoryContributor{}
	for rows.Next() {
		contributor := model.CategoryContributor{Categories: categories_params}
		err := rows.Scan(&contributor.LinksSubmitted, &contributor.LoginName)
		if err != nil {
			log.Fatal(err)
		}
		contributors = append(contributors, contributor)
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, contributors)
}

// GET TOP SUBCATEGORIES WITH GIVEN CATEGORY(IES)
func GetSubcategories(w http.ResponseWriter, r *http.Request) {
	categories_params := chi.URLParam(r, "categories")
	if categories_params == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("no categories provided")))
		return
	}
	// TODO: replace with middleware that converts all URLs to lowercase
	categories_params = strings.ToLower(categories_params)
	search_cats := strings.Split(categories_params, ",")
	
	get_subcats_sql := fmt.Sprintf(`select global_cats from Links where ',' || global_cats || ',' like '%%,%s,%%'`, search_cats[0])
	for i := 1; i < len(search_cats); i++ {
		get_subcats_sql += fmt.Sprintf(` AND ',' || global_cats || ',' like '%%,%s,%%'`, search_cats[i])
	}
	get_subcats_sql += ` group by global_cats;`

	db ,err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query(get_subcats_sql)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	
	var subcats []string
	for rows.Next() {
		var row_cats string
		err := rows.Scan(&row_cats)
		if err != nil {
			log.Fatal(err)
		}

		cats := strings.Split(row_cats, ",")
		for i := 0; i < len(cats); i++ {
			cat_lc := strings.ToLower(cats[i])

			// append to subcats if new and not in search_cats 
			if !slices.Contains(search_cats, cat_lc) && !slices.Contains(subcats, cat_lc) {
				subcats = append(subcats, cat_lc)
			}
		}
	}
	if len(subcats) == 0 {
		render.JSON(w, r, []string{})
		render.Status(r, http.StatusOK)
		return
	}

	subcats_with_counts := make([]model.CategoryCount, len(subcats))
	for i := 0; i < len(subcats); i++ {
		subcats_with_counts[i].Category = subcats[i]

		get_link_counts_sql := fmt.Sprintf(`SELECT count(*) as link_count FROM Links WHERE ',' || global_cats || ',' like '%%,%s,%%'`, subcats[i])

		for j := 0; j < len(search_cats); j++ {
			get_link_counts_sql += fmt.Sprintf(` AND ',' || global_cats || ',' LIKE '%%,%s,%%'`, search_cats[j])
		}
		get_link_counts_sql += `;`

		err := db.QueryRow(get_link_counts_sql).Scan(&subcats_with_counts[i].Count)
		if err != nil {
			log.Fatal(err)
		}
	}

	SortAndLimitCategoryCounts(&subcats_with_counts)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, subcats_with_counts)
}

func GetSubcategoriesByPeriod(w http.ResponseWriter, r *http.Request) {
	period_params, categories_params := chi.URLParam(r, "period"), chi.URLParam(r, "categories")
	if categories_params == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("no categories provided")))
		return
	} else if period_params == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("no period provided")))
		return
	}
	// TODO: replace with middleware that converts all URLs to lowercase
	categories_params = strings.ToLower(categories_params)
	search_cats := strings.Split(categories_params, ",")

	get_subcats_sql := `SELECT global_cats FROM Links`
	AppendPeriodClause(&get_subcats_sql, period_params)
	for _, cat := range search_cats {
		get_subcats_sql += fmt.Sprintf(` AND ',' || global_cats || ',' like '%%,%s,%%'`, cat)
	}
	get_subcats_sql += ` group by global_cats;`

	db ,err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query(get_subcats_sql)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	
	var subcats []string
	for rows.Next() {
		var row_cats string
		err := rows.Scan(&row_cats)
		if err != nil {
			log.Fatal(err)
		}

		cats := strings.Split(row_cats, ",")
		for i := 0; i < len(cats); i++ {
			cat_lc := strings.ToLower(cats[i])

			// append to subcats if new and not in search_cats 
			if !slices.Contains(search_cats, cat_lc) && !slices.Contains(subcats, cat_lc) {
				subcats = append(subcats, cat_lc)
			}
		}
	}
	if len(subcats) == 0 {
		render.JSON(w, r, []string{})
		render.Status(r, http.StatusOK)
		return
	}

	subcats_with_counts := make([]model.CategoryCount, len(subcats))
	for i := 0; i < len(subcats); i++ {
		subcats_with_counts[i].Category = subcats[i]

		get_link_counts_sql := `SELECT count(*) as link_count FROM Links`
		AppendPeriodClause(&get_link_counts_sql, period_params)
		get_link_counts_sql += fmt.Sprintf(` AND ',' || global_cats || ',' like '%%,%s,%%'`, subcats[i])
		for _, cat := range search_cats {
			get_link_counts_sql += fmt.Sprintf(` AND ',' || global_cats || ',' like '%%,%s,%%'`, cat)
		}
		get_link_counts_sql += `;`

		err := db.QueryRow(get_link_counts_sql).Scan(&subcats_with_counts[i].Count)
		if err != nil {
			log.Fatal(err)
		}
	}

	SortAndLimitCategoryCounts(&subcats_with_counts)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, subcats_with_counts)
}

func SortAndLimitCategoryCounts(cats_with_counts *[]model.CategoryCount) {
	// sort by count
	slices.SortFunc(*cats_with_counts, model.SortCategories)

	
	// limit to top {LIMIT} categories
	// 20 for now
	const LIMIT int = 20
	if len(*cats_with_counts) > LIMIT {
		*cats_with_counts = (*cats_with_counts)[:LIMIT]
	}
}

// ADD NEW LINK
func AddLink(w http.ResponseWriter, r *http.Request) {
	link_data := &model.NewLinkRequest{}
	if err := render.Bind(r, link_data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
	req_user_id, req_login_name, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	link_data.SubmittedBy = req_login_name

	// Check if more than 5 tag categories are submitted, Abort if so
	cat_limit := 5
	if strings.Count(link_data.NewLink.Categories, ",") > cat_limit {
		render.Render(w, r, ErrInvalidRequest(fmt.Errorf("tag has too many categories (%d max)", cat_limit)))
		return
	}

    // Check if URL contains http or https protocol, update if needed
	protocol_regex, err := regexp.Compile(`^(http(s?)\:\/\/)`)
	if err != nil {
		log.Fatal(err)
	}
	
	var resp *http.Response

	// Protocol not specified, try https then http
	if !protocol_regex.MatchString(link_data.NewLink.URL) {
		found := false

		// check https
		link_data.URL = "https://" + link_data.NewLink.URL
		resp, err = http.Get(link_data.URL)
		if err == nil {
			if resp.StatusCode > 299 && resp.StatusCode < 400 {
				render.Render(w, r, ErrInvalidRequest(errors.New("invalid link destination: redirect detected")))
				return
			} else {
				found = true
			}
		}

		// Check http if https not found
		if !found {
			link_data.URL = "http://" + link_data.NewLink.URL
			resp, err = http.Get(link_data.URL)
			if resp.StatusCode > 299 && resp.StatusCode < 400 {
				render.Render(w, r, ErrInvalidRequest(errors.New("invalid link destination: redirect detected")))
				return
			} else if err != nil {
				render.Render(w, r, ErrInvalidRequest(errors.New("invalid link: " + link_data.URL)))
				return
			}
		}

	// Protocol specified, check URL as-is
	} else {
		resp, err = http.Get(link_data.NewLink.URL)
		if err != nil || resp.StatusCode == 404 {
			render.Render(w, r, ErrInvalidRequest(errors.New("invalid link: " + link_data.URL)))
			return
		} else if resp.StatusCode > 299 && resp.StatusCode < 400 {
			render.Render(w, r, ErrInvalidRequest(errors.New("invalid link destination: redirect detected")))
			return
		}	
	}

	// Get full URL after any redirects e.g., to wwww.
	link_data.URL = resp.Request.URL.String()

	// Check if link already exists, Abort if attempting duplicate
	var s sql.NullString
	err = db.QueryRow("SELECT url FROM Links WHERE url = ?", link_data.URL).Scan(&s)
	if err == nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("link already exists")))
		return
	}

	// Extract meta tag content
	defer resp.Body.Close()
	meta := MetaFromHTMLTokens(resp.Body)

	// Get summary and summary author
	// (Auto-generate from meta tags if summary not provided)
	summary_author := ""

	// User-submitted summary
	if link_data.Summary != "" {
		summary_author = req_user_id

	// No user-submitted: use Auto Summary
	} else {
		auto_summary := ""

		switch {
			case meta.OGDescription != "":
				auto_summary = meta.OGDescription
			case meta.Description != "":
				auto_summary = meta.Description
			case meta.OGTitle != "":
				auto_summary = meta.OGTitle
			case meta.Title != "":
				auto_summary = meta.Title
			case meta.OGSiteName != "":
				auto_summary = meta.OGSiteName
		}

		// Auto Summary successfully retrieved: assign to request
		if auto_summary != "" {
			link_data.Summary = auto_summary

			// 15 is Auto Summary's user_id
			// TODO: update with final
			summary_author = "15"
		
		// Else no summary (sad!)
		} else {
			summary_author = ""
		}
	}

	// Get og:image, if available, for link preview
	var og_image string
	if meta.OGImage != "" {

		// check that image link is valid
		resp, err := http.Get(meta.OGImage)
		if err != nil || resp.StatusCode == 404 || (resp.StatusCode > 299 && resp.StatusCode < 400) {

			// use no image if link is invalid
			og_image = ""
		} else {
			og_image = meta.OGImage
		}
	}
	link_data.ImgURL = og_image

	// Sort categories alphabetically
	split_categories := strings.Split(link_data.NewLink.Categories, ",")
	slices.Sort(split_categories)
	link_data.Categories = strings.Join(split_categories, ",")

	// Insert link
	res, err := db.Exec("INSERT INTO Links VALUES(?,?,?,?,?,?,?);", nil, link_data.URL, req_login_name, link_data.SubmitDate, link_data.Categories, link_data.Summary, og_image)
	if err != nil {
		log.Fatal(err)
	}

	// Get new link ID
	var id int64
	if id, err = res.LastInsertId(); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
	}
	link_data.ID = id

	// Create new summary if retrieved from request or auto_summary
	if link_data.Summary != "" {
		_, err = db.Exec("INSERT INTO Summaries VALUES(?,?,?,?,?);", nil, link_data.Summary, link_data.ID, summary_author,link_data.SubmitDate)
		if err != nil {
			log.Fatal(err)
		}

		link_data.SummaryCount = 1
	}

	// Create initial tag
	_, err = db.Exec("INSERT INTO Tags VALUES(?,?,?,?,?);", nil, link_data.ID, link_data.Categories, req_login_name, link_data.SubmitDate)
	if err != nil {
		log.Fatal(err)
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, link_data)
}

// LIKE LINK
func LikeLink(w http.ResponseWriter, r *http.Request) {
	link_id := chi.URLParam(r, "link_id")
	if link_id == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("invalid link ID provided")))
		return
	}

	req_user_id, _, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check if link doesn't exist or if link submitted by same user, Abort if either
	var link_submitted_by_name sql.NullString
	err = db.QueryRow("SELECT submitted_by FROM Links WHERE id = ?;", link_id).Scan(&link_submitted_by_name)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("invalid link ID")))
		return
	}

	var link_submitted_by_id sql.NullInt64
	err = db.QueryRow("SELECT id FROM Users WHERE login_name = ?;",link_submitted_by_name.String).Scan(&link_submitted_by_id)
	if err != nil {
		log.Fatal(err)
	}

	req_user_id_int64, err := strconv.ParseInt(req_user_id, 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	if link_submitted_by_id.Int64 == req_user_id_int64 {
		render.Render(w, r, ErrInvalidRequest(errors.New("cannot like your own link")))
		return
	}

	// Check if user already liked this link, Abort if already liked
	var l sql.NullString
	err = db.QueryRow("SELECT id FROM 'Link Likes' WHERE link_id = ? AND user_id = ?;", link_id, req_user_id).Scan(&l)
	if err == nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("already liked")))
		return
	}

	res, err := db.Exec("INSERT INTO 'Link Likes' VALUES(?,?,?);", nil, req_user_id, link_id)
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
		render.Render(w, r, ErrInvalidRequest(errors.New("invalid link ID provided")))
		return
	}

	req_user_id, _, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check if link like submitted by requesting user exists, Abort if not
	var like_id sql.NullString
	err = db.QueryRow("SELECT id FROM 'Link Likes' WHERE user_id = ? AND link_id = ?;", req_user_id, link_id).Scan(&like_id)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("link like not found")))
		return
	}

	// Delete like
	_, err = db.Exec("DELETE FROM 'Link Likes' WHERE id = ?;", like_id)
	if err != nil {
		log.Fatal(err)
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{"message": "deleted"})
}

// COPY LINK TO USER'S TREASURE MAP
func CopyLink(w http.ResponseWriter, r *http.Request) {
	link_id := chi.URLParam(r, "link_id")
	if link_id == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("no link ID provided")))
		return
	}

	req_user_id, _, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check if link already in map, Abort if attempting duplicate
	var l sql.NullString
	err = db.QueryRow("SELECT id FROM 'Link Copies' WHERE link_id = ? AND user_id = ?;", link_id, req_user_id).Scan(&l)
	if err == nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("link already in map")))
		return
	}

	res, err := db.Exec("INSERT INTO 'Link Copies' VALUES(?,?,?);", nil, link_id, req_user_id)
	if err != nil {
		log.Fatal(err)
	}

	// Get ID of new link copy
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
		render.Render(w, r, ErrInvalidRequest(errors.New("no link ID provided")))
		return
	}

	req_user_id, _, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check if link copy exists and was submitted by same user, Abort if either unsatisfied
	var cid sql.NullString
	err = db.QueryRow("SELECT id FROM 'Link Copies' WHERE link_id = ? AND user_id = ?;", link_id, req_user_id).Scan(&cid)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("link copy does not exist")))
		return
	}

	// Delete
	_, err = db.Exec("DELETE FROM 'Link Copies' WHERE id = ?;", cid.String)
	if err != nil {
		log.Fatal(err)
	}

	return_json := map[string]string{
		"message": "deleted",
	}

	render.JSON(w, r, return_json)
	render.Status(r, http.StatusNoContent)
}

// GET LINK LIKES
// (not currently used - probably delete later)
func GetLinkLikes(w http.ResponseWriter, r *http.Request) {
	link_id := chi.URLParam(r, "id")
	if link_id == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("invalid link id provided")))
		return
	}

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	// Check if link exists, Abort if invalid link ID provided
	var s sql.NullString
	err = db.QueryRow("SELECT id FROM Links WHERE id = ?;", link_id).Scan(&s)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("no link found with given ID")))
		return
	}

	// Get like count
	var c int64
	err = db.QueryRow("SELECT COUNT(id) as count FROM 'Link Likes' WHERE link_id = ?;", link_id).Scan(&c)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	return_json := map[string]int64{"likes": c}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, return_json)
}

func ScanLinksSignedOut(db *sql.DB, rows *sql.Rows) *[]model.LinkSignedOut {
	var links = []model.LinkSignedOut{}

	for rows.Next() {
		i := model.LinkSignedOut{}
		err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.Categories, &i.Summary, &i.SummaryCount, &i.LikeCount, &i.ImgURL)
		if err != nil {
			log.Fatal(err)
		}

		links = append(links, i)
	}

	return &links
}

func ScanLinksSignedIn(db *sql.DB, rows *sql.Rows, user_id string) *[]model.LinkSignedIn {
	var links = []model.LinkSignedIn{}

	for rows.Next() {
		i := model.LinkSignedIn{}
		err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.Categories, &i.Summary, &i.SummaryCount, &i.LikeCount, &i.ImgURL)
		if err != nil {
			log.Fatal(err)
		}

		// Add IsLiked / IsCopied / IsTagged
		var l sql.NullInt32
		var t sql.NullInt32
		var c sql.NullInt32

		err = db.QueryRow(fmt.Sprintf(`SELECT
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
		) as is_copied;`, i.ID, user_id)).Scan(&l,&t, &c)
		if err != nil {
			log.Fatal(err)
		}

		i.IsLiked = l.Int32 > 0
		i.IsTagged = t.Int32 > 0
		i.IsCopied = c.Int32 > 0

		links = append(links, i)
	}
	return &links
}