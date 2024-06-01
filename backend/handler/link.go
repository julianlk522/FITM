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
	"github.com/jonlaing/htmlmeta"
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
	get_link_likes_sql := fmt.Sprintf(`SELECT links_id as link_id, url, link_author as submitted_by, submit_date, categories, summary, coalesce(count(Summaries.id),0) as summary_count, like_count, img_url
	FROM (SELECT Links.id as links_id, url, submitted_by as link_author, submit_date, coalesce(global_cats,"") as categories, coalesce(global_summary,"") as summary, coalesce(like_count,0) as like_count, coalesce(img_url,"") as img_url FROM LINKS LEFT JOIN (SELECT link_id as likes_link_id, count(*) as like_count FROM 'Link Likes' GROUP BY likes_link_id) ON Links.id = likes_link_id) LEFT JOIN Summaries ON Summaries.link_id = links_id GROUP BY links_id ORDER BY like_count DESC, links_id ASC LIMIT %s;`, LIMIT)
	rows, err := db.Query(get_link_likes_sql)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Check auth token
	var req_user_id string
	claims, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if len(claims) > 0 {
		req_user_id = claims["user_id"].(string)
	}

	// Scan links
	// User signed in: get isLiked for each link
	if req_user_id != "" {
		var links []model.LinkSignedIn
		for rows.Next() {
			i := model.LinkSignedIn{}
			err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.Categories, &i.Summary, &i.SummaryCount, &i.LikeCount, &i.ImgURL)
			if err != nil {
				log.Fatal(err)
			}
	
			// Add IsLiked
			var l sql.NullInt32
			err = db.QueryRow("SELECT count(*) as is_liked FROM 'Link Likes' WHERE link_id = ? AND user_id = ?;", i.ID, req_user_id).Scan(&l)
			if err != nil {
				log.Fatal(err)
			}
			i.IsLiked = l.Int32 > 0
			links = append(links, i)
		}
		render.JSON(w, r, links)
		
	// User not signed in: omit isLiked
	} else {
		var links []model.Link
		for rows.Next() {
			i := model.Link{}
			err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.Categories, &i.Summary, &i.SummaryCount, &i.LikeCount, &i.ImgURL)
			if err != nil {
				log.Fatal(err)
			}
			links = append(links, i)
		}
		render.JSON(w, r, links)
	}

	render.Status(r, http.StatusOK)
}

// GET MOST-LIKED LINKS DURING PERIOD
// (day, week, month)
// (top 20 for now)
func GetTopLinksByPeriod(w http.ResponseWriter, r *http.Request) {
	db ,err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	get_link_likes_sql := `SELECT links_id as link_id, url, link_author as subitted_by, submit_date, categories, summary, coalesce(count(Summaries.id),0) as summary_count, like_count, img_url
	FROM
		(
		SELECT Links.id as links_id, url, submitted_by as link_author, submit_date, coalesce(global_cats,"") as categories, coalesce(global_summary,"") as summary, coalesce(like_count,0) as like_count, coalesce(img_url,"") as img_url
		FROM LINKS
		LEFT JOIN 
			(
			SELECT link_id as likes_link_id, count(*) as like_count
			FROM 'Link Likes'
			GROUP BY likes_link_id
			)
		ON Links.id = likes_link_id`

	switch chi.URLParam(r, "period") {
	case "day":
		get_link_likes_sql += ` WHERE julianday('now') - julianday(submit_date) <= 2`
	case "week":
		get_link_likes_sql += ` WHERE julianday('now') - julianday(submit_date) <= 8`
	case "month":
		get_link_likes_sql += ` WHERE julianday('now') - julianday(submit_date) <= 31`
	default:
		render.Render(w, r, ErrInvalidRequest(errors.New("invalid period")))
		return
	}

	const LIMIT string = "20"
	get_link_likes_sql += fmt.Sprintf(`) LEFT JOIN Summaries
	ON Summaries.link_id = links_id
	GROUP BY links_id ORDER BY like_count DESC, link_id ASC LIMIT %s;`, LIMIT)

	rows, err := db.Query(get_link_likes_sql)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Check auth token
	var req_user_id string
	claims, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if len(claims) > 0 {
		req_user_id = claims["user_id"].(string)
	}
	
	// Scan links
	// User signed in: get isLiked for each link
	if req_user_id != "" {
		var links []model.LinkSignedIn
		for rows.Next() {
			i := model.LinkSignedIn{}
			err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.Categories, &i.Summary, &i.SummaryCount, &i.LikeCount, &i.ImgURL)
			if err != nil {
				log.Fatal(err)
			}
	
			// Add IsLiked
			var l sql.NullInt32
			err = db.QueryRow("SELECT count(*) as is_liked FROM 'Link Likes' WHERE link_id = ? AND user_id = ?;", i.ID, req_user_id).Scan(&l)
			if err != nil {
				log.Fatal(err)
			}
			i.IsLiked = l.Int32 > 0
			links = append(links, i)
		}
		render.JSON(w, r, links)
		
	// User not signed in: omit isLiked
	} else {
		var links []model.Link
		for rows.Next() {
			i := model.Link{}
			err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.Categories, &i.Summary, &i.SummaryCount, &i.LikeCount, &i.ImgURL)
			if err != nil {
				log.Fatal(err)
			}
			links = append(links, i)
		}
		render.JSON(w, r, links)
	}

	render.Status(r, http.StatusOK)
}

// GET MOST-LIKED LINKS WITH GIVEN CATEGORY(IES)
// (top 20 for now)
func GetTopLinksByCategories(w http.ResponseWriter, r *http.Request) {
	db ,err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// get categories
	categories_params := chi.URLParam(r, "categories")
	var get_links_sql string

	// multiple categories
	if strings.Contains(categories_params, ",") {
		categories := strings.Split(categories_params, ",")

		// get link IDs
		get_links_sql = fmt.Sprintf(`select id from Links where ',' || global_cats || ',' like '%%,%s,%%'`, categories[0])

		for i := 1; i < len(categories); i++ {
			get_links_sql += fmt.Sprintf(` AND ',' || global_cats || ',' like '%%,%s,%%'`, categories[i])
		}

	// single category
	} else {

		// get link IDs
		get_links_sql = fmt.Sprintf(`select id from Links where ',' || global_cats || ',' like '%%,%s,%%'`, categories_params)
	}
	get_links_sql += ` group by id`

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

	// if no links found, return status message
	if len(link_ids) == 0 {
		render.JSON(w, r, []model.Link{})
		render.Status(r, http.StatusOK)
		return
	}

	// get link data for each ID
	db, err = sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	const LIMIT string = "20"
	rows, err = db.Query(fmt.Sprintf(`SELECT links_id as link_id, url, link_author as submitted_by, submit_date, categories, summary, coalesce(count(Summaries.id),0) as summary_count, like_count, img_url
	FROM
		(
		SELECT Links.id as links_id, url, submitted_by as link_author, submit_date, coalesce(global_cats,"") as categories, coalesce(global_summary,"") as summary, coalesce(like_count,0) as like_count, coalesce(img_url,"") as img_url
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

	// Check auth token
	var req_user_id string
	claims, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if len(claims) > 0 {
		req_user_id = claims["user_id"].(string)
	}

	// Scan links
	// User signed in: get isLiked for each link
	if req_user_id != "" {
		var links []model.LinkSignedIn
		for rows.Next() {
			i := model.LinkSignedIn{}
			err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.Categories, &i.Summary, &i.SummaryCount, &i.LikeCount, &i.ImgURL)
			if err != nil {
				log.Fatal(err)
			}
	
			// Add IsLiked
			var l sql.NullInt32
			err = db.QueryRow("SELECT count(*) as is_liked FROM 'Link Likes' WHERE link_id = ? AND user_id = ?;", i.ID, req_user_id).Scan(&l)
			if err != nil {
				log.Fatal(err)
			}
			i.IsLiked = l.Int32 > 0
			links = append(links, i)
		}
		render.JSON(w, r, links)
		
	// User not signed in: omit isLiked
	} else {
		var links []model.Link
		for rows.Next() {
			i := model.Link{}
			err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.Categories, &i.Summary, &i.SummaryCount, &i.LikeCount, &i.ImgURL)
			if err != nil {
				log.Fatal(err)
			}
			links = append(links, i)
		}
		render.JSON(w, r, links)
	}	

	render.Status(r, http.StatusOK)
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
		var contributor model.CategoryContributor
		contributor.Categories = categories_params
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

// todo: change from Tags (categories) to Links (global_cats) once there is more data to query
func GetTopSubcategories(w http.ResponseWriter, r *http.Request) {

	// Limit 20 for now
	const LIMIT int = 20

	db ,err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// get categories
	search_cats_params := chi.URLParam(r, "categories")
	// todo: replace with middleware that converts all URLs to lowercase
	search_cats_params = strings.ToLower(search_cats_params)
	search_cats := strings.Split(search_cats_params, ",")
	
	// get subcategories
	get_links_sql := fmt.Sprintf(`select categories from Tags where ',' || categories || ',' like '%%,%s,%%'`, search_cats[0])
	for i := 1; i < len(search_cats); i++ {
		get_links_sql += fmt.Sprintf(` AND ',' || categories || ',' like '%%,%s,%%'`, search_cats[i])
	}
	get_links_sql += ` group by categories;`

	rows, err := db.Query(get_links_sql)
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
			if !slices.Contains(search_cats, cat_lc) && !slices.Contains(subcats, cat_lc) {
				subcats = append(subcats, cat_lc)
			}
		}
	}

	// if no links found, return status message
	if len(subcats) == 0 {
		render.JSON(w, r, []string{})
		render.Status(r, http.StatusOK)
		return
	}

	// get total links for each subcategory
	subcats_with_counts := make([]model.CategoryCount, len(subcats))
	for i := 0; i < len(subcats); i++ {
		subcats_with_counts[i].Category = subcats[i]

		get_link_counts_sql := fmt.Sprintf(`select count(*) as link_count from Tags where ',' || categories || ',' like '%%,%s,%%'`, subcats[i])

		for j := 0; j < len(search_cats); j++ {
			get_link_counts_sql += fmt.Sprintf(` AND ',' || categories || ',' like '%%,%s,%%'`, search_cats[j])
		}
		get_link_counts_sql += `;`

		err := db.QueryRow(get_link_counts_sql).Scan(&subcats_with_counts[i].Count)
		if err != nil {
			log.Fatal(err)
		}
	}

	// sort by count
	slices.SortFunc(subcats_with_counts, model.SortCategories)

	// limit to top {LIMIT} categories
	if len(subcats_with_counts) > LIMIT {
		subcats_with_counts = subcats_with_counts[0:LIMIT]
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, subcats_with_counts)
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

	// Check auth token
	var req_user_id, req_login_name string
	claims, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if len(claims) > 0 {
		req_user_id = claims["user_id"].(string)
		req_login_name = claims["login_name"].(string)

		link_data.SubmittedBy = req_login_name
	}

	// Check if link contains any subdomains
	regex, _ := regexp.Compile(`^(?:http[s]?\:\/\/)?(?:[^\/\W]+?\.){2,}(?:[^\/\n\r]+)`)
	// regex should match:
	// www.google.com
	// www.www.google.com
	// https://www.google.com
	// http://www.google.com
	// http://www.www.google.com
	// etc.

	// should not match:
	// google.com
	// https://google.com
	// etc.
	subdomain_found := regex.MatchString(link_data.URL)
	if !subdomain_found {

		// Prepend "https://www."" if no subdomain or protocol found
		if !strings.HasPrefix(link_data.URL, "https://") {
			link_data.URL = "https://www." + link_data.URL
		
		// Else append "www." after "https://" if protocol found but not subdomain
		} else {
			link_data.URL = strings.Replace(link_data.URL, "https://", "https://www.", 1)
		}
		
	// Else prepend "https://" if subdomain found but protocol not
	} else if (!strings.HasPrefix(link_data.URL, "http")) {
		link_data.URL = "https://" + link_data.URL
	}

	// Check if link exists, Abort if attempting duplicate
	var s sql.NullString
	err = db.QueryRow("SELECT url FROM Links WHERE url = ?", link_data.URL).Scan(&s)
	if err == nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("link already exists")))
		return
	}

	// Verify that link is valid
	resp, err := http.Get(link_data.URL)
	if err != nil || resp.StatusCode == 404 {
		render.Render(w, r, ErrInvalidRequest(errors.New("invalid link: " + link_data.URL)))
		return
	} else if resp.StatusCode > 299 && resp.StatusCode < 400 {
		log.Println(resp)
		render.Render(w, r, ErrInvalidRequest(errors.New("invalid link destination: redirect detected")))
		return
	}

	// Extract meta data
	defer resp.Body.Close()
	meta := htmlmeta.Extract(resp.Body)

	// Get automatically-generated link summary from meta title or description
	var summary_count int = 1
	var auto_summary string
	if meta.Title != "" {
		auto_summary = meta.Title
	} else if meta.Description != "" {
		auto_summary = meta.Description
	} else if meta.OGDescription != "" {
		auto_summary = meta.OGDescription
	} else {
		// no extractible summary
		summary_count = 0
	}

	link_data.Summary = auto_summary
	link_data.SummaryCount = summary_count

	// Get og:image, if available, for link preview image
	var og_image string
	if meta.OGImage != "" {
		og_image = meta.OGImage
	}
	link_data.ImgURL = og_image

	res, err := db.Exec("INSERT INTO Links VALUES(?,?,?,?,?,?,?);", nil, link_data.URL, req_login_name, link_data.SubmitDate, "", auto_summary, og_image)
	if err != nil {
		log.Fatal(err)
	}

	// Create new summary if auto_summary successfully retrieves a title or description
	if auto_summary != "" {
		_, err = db.Exec("INSERT INTO Summaries VALUES(?,?,?,?);", nil, auto_summary, link_data.ID, req_user_id)
		if err != nil {
			log.Fatal(err)
		}	
	}

	var id int64
	if id, err = res.LastInsertId(); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
	}
	link_data.ID = id

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

	// Check auth token
	var req_user_id string
	claims, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if len(claims) > 0 {
		req_user_id = claims["user_id"].(string)
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
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, like_link_data)
}

// UN-LIKE LINK
func UnlikeLink(w http.ResponseWriter, r *http.Request) {
	link_id := chi.URLParam(r, "link_id")
	if link_id == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("invalid link ID provided")))
		return
	}

	// Check auth token
	var req_user_id string
	claims, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if len(claims) > 0 {
		req_user_id = claims["user_id"].(string)
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
func CopyLinkToMap(w http.ResponseWriter, r *http.Request) {
	copy_link_data := &model.LinkCopyRequest{}
	if err := render.Bind(r, copy_link_data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Check auth token
	var req_user_id string
	claims, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if len(claims) > 0 {
		req_user_id = claims["user_id"].(string)
	}

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check if link already in map, Abort if attempting duplicate
	var l sql.NullString
	err = db.QueryRow("SELECT id FROM 'Link Copies' WHERE link_id = ? AND user_id = ?;", copy_link_data.LinkID, req_user_id).Scan(&l)
	if err == nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("link already in map")))
		return
	}

	res, err := db.Exec("INSERT INTO 'Link Copies' VALUES(?,?,?);", nil, copy_link_data.LinkID, req_user_id)
	if err != nil {
		log.Fatal(err)
	}

	var id int64
	if id, err = res.LastInsertId(); err != nil {
		log.Fatal(err)
	}

	return_json := map[string]int64{
		"copy_id": id,
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, return_json)
}

// UN-COPY LINK
func UncopyLink(w http.ResponseWriter, r *http.Request) {
	delete_copy_data := &model.DeleteLinkCopyRequest{}
	if err := render.Bind(r, delete_copy_data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Check auth token
	var req_user_id string
	claims, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if len(claims) > 0 {
		req_user_id = claims["user_id"].(string)
	}

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check if link copy exists and was submitted by same user, Abort if either unsatisfied
	var s sql.NullString
	err = db.QueryRow("SELECT id FROM 'Link Copies' WHERE id = ? AND user_id = ?;", delete_copy_data.ID, req_user_id).Scan(&s)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("link copy does not exist")))
		return
	}

	// Delete
	_, err = db.Exec("DELETE FROM 'Link Copies' WHERE id = ?;", delete_copy_data.ID)
	if err != nil {
		log.Fatal(err)
	}

	return_json := map[string]string{
		"status": "success",
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