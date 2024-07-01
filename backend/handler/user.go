package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"golang.org/x/crypto/bcrypt"

	"oitm/model"
)

var pic_dir string

func init() {
	work_dir, _ := os.Getwd()
	pic_dir = filepath.Join(work_dir, "db/profile-pics")
}

// SIGN UP
func SignUp(w http.ResponseWriter, r *http.Request) {
	signup_data := &model.SignUpRequest{}

	if err := render.Bind(r, signup_data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check if user already exists, Abort if so
	var s sql.NullString
	err = db.QueryRow("SELECT login_name FROM Users WHERE login_name = ?", signup_data.UserAuth.LoginName).Scan(&s)
	if err == nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("login name taken")))
		return
	}

	// Hash password
	pw_hash, err := bcrypt.GenerateFromPassword([]byte(signup_data.UserAuth.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`INSERT INTO users VALUES (?,?,?,?,?,?)`, nil, signup_data.UserAuth.LoginName, pw_hash, nil, nil, signup_data.Created)
	if err != nil {
		log.Fatal(err)
	}

	// get new user ID
	var id sql.NullString
	err = db.QueryRow("SELECT id FROM Users WHERE login_name = ?", signup_data.UserAuth.LoginName).Scan(&id)
	if err != nil {
		log.Fatal(err)
	}

	// generate and return jwt containing user ID and login_name
	token_data := map[string]interface{}{"user_id": id.String, "login_name": signup_data.LoginName}
	token_auth := jwtauth.New("HS256", []byte("secret"), nil, jwt.WithAcceptableSkew(24*time.Hour))
	_, token, err := token_auth.Encode(token_data)
	if err != nil {
		log.Fatal(err)
	}

	return_json := map[string]string{"token": token}
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, return_json)
}

// LOG IN
func LogIn(w http.ResponseWriter, r *http.Request) {
	login_data := &model.LogInRequest{}

	if err := render.Bind(r, login_data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Attempt to collect user ID and hashed password, 
	// Abort if user not found
	var id, p sql.NullString
	err = db.QueryRow("SELECT id, password FROM Users WHERE login_name = ?", login_data.LoginName).Scan(&id, &p)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("no user found with given login name")))
		return
	}

	// compare password hashes
	err = bcrypt.CompareHashAndPassword([]byte(p.String), []byte(login_data.UserAuth.Password))
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("incorrect password")))
		return
	}

	// generate and return jwt containing user ID and login_name
	token_data := map[string]interface{}{"user_id": id.String, "login_name": login_data.LoginName}
	token_auth := jwtauth.New("HS256", []byte("secret"), nil, jwt.WithAcceptableSkew(24*time.Hour))
	_, token, err := token_auth.Encode(token_data)
	if err != nil {
		log.Fatal(err)
	}

	return_json := map[string]string{"token": token}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, return_json)
}

// GET PROFILE
func GetProfile(w http.ResponseWriter, r *http.Request) {
	login_name := chi.URLParam(r, "login_name")
	if login_name == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("invalid login name provided")))
		return
	}

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var u model.User
	err = db.QueryRow(`SELECT login_name, coalesce(about,"") as about, coalesce(pfp,"") as pfp, coalesce(created,"") as created FROM Users WHERE login_name = ?;`, login_name).Scan(&u.LoginName, &u.About, &u.PFP, &u.Created)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("user not found")))
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, u)
}

// EDIT PROFILE
func EditProfile(w http.ResponseWriter, r *http.Request) {
	edit_profile_data := &model.EditProfileRequest{}

	if err := render.Bind(r, edit_profile_data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
	// Check auth token
	var req_user_id string
	claims, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if len(claims) > 0 {
		req_user_id = claims["user_id"].(string)
	}

	// Update profile
	return_json := map[string]interface{}{}

	// About
	if edit_profile_data.EditAboutRequest != nil {
		// TODO replace hard-coded id with id corresponding
		// to provided auth token
		_, err = db.Exec(`UPDATE Users SET about = ? WHERE id = ?`, edit_profile_data.EditAboutRequest.About, req_user_id)
		if err != nil {
			log.Fatal(err)
		}

		return_json["about"] = edit_profile_data.EditAboutRequest.About
	}
	
	// Profile Pic
	if edit_profile_data.EditPfpRequest != nil {
		// TODO replace hard-coded id with id corresponding
		// to provided auth token
		_, err = db.Exec(`UPDATE Users SET pfp = ? WHERE id = ?`, edit_profile_data.EditPfpRequest.PFP, req_user_id)
		if err != nil {
			log.Fatal(err)
		}

		return_json["pfp"] = edit_profile_data.EditPfpRequest.PFP
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, return_json)
}

// GET PROFILE PICTURE
// (from backend/db/profile-pics/{file_name})
func GetProfilePic(w http.ResponseWriter, r *http.Request) {
	var file_name string = chi.URLParam(r, "file_name")
	path := pic_dir + "/" + file_name
	
	// Error if pic not found at path
	if _, err := os.Stat(path); err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("profile pic not found")))
		return
	}
	
	// Serve if found
	http.ServeFile(w, r, path)
}

// UPLOAD NEW PROFILE PICTURE
func UploadProfilePic(w http.ResponseWriter, r *http.Request) {

	// Get file up to 10MB
	r.ParseMultipartForm( 10 << 20 )
	file, handler, err := r.FormFile("pic")
    if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
        return
    }
    defer file.Close()

	// Check that file is valid image
	if !strings.Contains(handler.Header.Get("Content-Type"), "image") {
		render.Render(w, r, ErrInvalidRequest(errors.New("invalid file provided")))
		return
	}

	// Get file extension
	ext := filepath.Ext(handler.Filename)

	// Generate unique file name
	new_name := uuid.New().String() + ext
	full_path := pic_dir + "/" + new_name

	// Create new file
	dst, err := os.Create(full_path)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("could not create new file")))
		return
	}
	defer dst.Close()

	// Save to new file
	if _, err := io.Copy(dst, file); err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("could not copy profile pic into new file")))
		return
	}

	// Get requesting user from auth token context
	var req_user_id string
	claims, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if len(claims) > 0 {
		req_user_id = claims["user_id"].(string)
	}

	// Update db with new pic name
	db, err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
	_, err = db.Exec(`UPDATE Users SET pfp = ? WHERE id = ?`, new_name, req_user_id)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("could not save new profile pic")))
		return
	}

	// Return saved file
	http.ServeFile(w, r, full_path)
}

// GET USER TREASURE MAP
// (includes links tagged by and copied by user, plus category sum counts)
// (all links submitted by a user will have a tag from that user, so includes all user's submitted links)
func GetTreasureMap(w http.ResponseWriter, r *http.Request) {
	var login_name string = chi.URLParam(r, "login_name")
	if login_name == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("no user provided")))
		return
	}

	db ,err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// check that user exists
	var u sql.NullString
	err = db.QueryRow("SELECT login_name FROM Users WHERE login_name = ?;", login_name).Scan(&u)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("user not found")))
		return
	}

	// Check auth token
	var req_user_id, req_login_name string
	claims, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if len(claims) > 0 {
		req_user_id = claims["user_id"].(string)
		req_login_name = claims["login_name"].(string)
	}

	// Prepare SQL to get submitted / tagged / copied links from User
	// (Start with queries for signed-out user, append additional if needed)
	base_fields := `SELECT 
		Links.id as link_id, 
		url, 
		submitted_by as login_name, 
		submit_date, 
		categories, 
		coalesce(global_summary,"") as summary, 
		coalesce(summary_count,0) as summary_count, 
		coalesce(like_count,0) as like_count, 
		coalesce(img_url,"") as img_url
	`

	// Get submitted links, replacing global categories with user-assigned
	submitted_from := fmt.Sprintf(` FROM Links
	JOIN
		(
		SELECT categories, link_id as tag_link_id
		FROM Tags
		WHERE submitted_by = '%s'
		)
	ON link_id = tag_link_id
	LEFT JOIN
		(
		SELECT count(*) as like_count, link_id as like_link_id
		FROM 'Link Likes'
		GROUP BY link_id
		)
	ON like_link_id = link_id
	LEFT JOIN
		(
		SELECT count(*) as summary_count, link_id as summary_link_id
		FROM Summaries
		GROUP BY link_id
		)
	ON summary_link_id = link_id`, login_name)

	submitted_where := fmt.Sprintf(` WHERE submitted_by = '%s';`, login_name)

	// Get tagged links submitted by other users, replacing global categories with user-assigned
	tagged_from := submitted_from
	tagged_where := fmt.Sprintf(` WHERE submitted_by != '%s';`, login_name)

	// Get copied links
	copied_fields := strings.Replace(base_fields, "categories", `coalesce(global_cats,"") as categories`, 1)
	copied_from := fmt.Sprintf(` FROM Links
	JOIN
		(
		SELECT link_id as copy_link_id, user_id as copier_id
		FROM 'Link Copies'
		JOIN Users
		ON Users.id = copier_id
		WHERE Users.login_name = '%s'
		)
	ON copy_link_id = link_id
	LEFT JOIN
		(
		SELECT count(*) as like_count, link_id as like_link_id
		FROM 'Link Likes'
		GROUP BY link_id
		)
	ON like_link_id = link_id
	LEFT JOIN
		(
		SELECT count(*) as summary_count, link_id as summary_link_id
		FROM Summaries
		GROUP BY link_id
		)
	ON summary_link_id = link_id`, login_name)

	// Append additional queries for signed-in fields (IsLiked, IsTagged, IsCopied) if auth claims verified
	if req_user_id != "" {
		added_fields := `, 
		coalesce(is_liked,0) as is_liked, 
		coalesce(is_tagged,0) as is_tagged,
		coalesce(is_copied,0) as is_copied`

		added_from := fmt.Sprintf(` LEFT JOIN
			(
			SELECT id, count(*) as is_liked, user_id, link_id as like_link_id2
			FROM 'Link Likes'
			WHERE user_id = '%[1]s'
			GROUP BY id
			)
		ON like_link_id2 = link_id 
		LEFT JOIN 
		(
			SELECT id as tag_id, link_id as tlink_id, count(*) as is_tagged 
			FROM Tags
			WHERE Tags.submitted_by = '%[2]s'
			GROUP BY tag_id
		)
		ON tlink_id = link_id
		LEFT JOIN
			(
			SELECT id as copy_id, count(*) as is_copied, user_id as cuser_id, link_id as clink_id
			FROM 'Link Copies'
			WHERE cuser_id = '%[1]s'
			GROUP BY copy_id
			)
		ON clink_id = link_id`, req_user_id, req_login_name)

		// Submitted
		submitted_fields := base_fields + added_fields
		submitted_from += added_from
		submitted_sql := submitted_fields + submitted_from + submitted_where

		// Tagged
		tagged_fields := base_fields + added_fields
		tagged_from += added_from
		tagged_sql := tagged_fields + tagged_from + tagged_where

		// Copied
		copied_fields += added_fields
		copied_from += added_from
		copied_where := fmt.Sprintf(` WHERE link_id NOT IN
			(
			SELECT link_id
			FROM TAGS
			WHERE submitted_by = '%s'
			);`, login_name)
		copied_sql := copied_fields + copied_from + copied_where

		// Scan links
		var submitted, tagged, copied *[]model.LinkSignedIn
		for _, sql := range []string{submitted_sql, tagged_sql, copied_sql} {
			rows, err := db.Query(sql)
			if err != nil {
				log.Fatal(err)
			}
			defer rows.Close()
			
			switch sql {
				case submitted_sql:
					submitted = ScanTmapLinksSignedIn(db, rows)
				case tagged_sql:
					tagged = ScanTmapLinksSignedIn(db, rows)
				case copied_sql:
					copied = ScanTmapLinksSignedIn(db, rows)
			}
		}

		// Add links to tmap
		tmap := model.TreasureMap[model.LinkSignedIn]{Submitted: submitted, Tagged: tagged, Copied: copied}

		// Get category counts
		all_links := slices.Concat(*submitted, *tagged, *copied)
		cat_counts := GetTmapCategoryCounts(&all_links, nil)

		// combine links and categories in response
		tmap.Categories = cat_counts
		render.JSON(w, r, tmap)
		
	// User not signed in: omit isLiked / isCopied / isTagged fields
	} else {
		get_submitted_sql := base_fields + submitted_from + submitted_where
		get_tagged_sql := base_fields + tagged_from + tagged_where
		get_copied_sql := copied_fields + copied_from

		// Scan links
		var submitted, tagged, copied *[]model.LinkSignedOut
		for _, sql := range []string{get_submitted_sql, get_tagged_sql, get_copied_sql} {
			rows, err := db.Query(sql)
			if err != nil {
				log.Fatal(err)
			}
			defer rows.Close()
			
			switch sql {
			case get_submitted_sql:
				submitted = ScanTmapLinksSignedOut(db, rows)
			case get_tagged_sql:
				tagged = ScanTmapLinksSignedOut(db, rows)
			case get_copied_sql:
				copied = ScanTmapLinksSignedOut(db, rows)
			}
		}

		// Add links to tmap
		tmap := model.TreasureMap[model.LinkSignedOut]{Submitted: submitted, Tagged: tagged, Copied: copied}

		// Get category counts
		all_links := slices.Concat(*submitted, *tagged, *copied)
		cat_counts := GetTmapCategoryCounts(&all_links, nil)

		// Add categories to tmap
		tmap.Categories = cat_counts
		render.JSON(w, r, tmap)
	}
}

func GetTreasureMapByCategories(w http.ResponseWriter, r *http.Request) {
	var login_name string = chi.URLParam(r, "login_name")
	if login_name == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("no user provided")))
		return
	}

	var categories string = chi.URLParam(r, "categories")
	if categories == "" {
		render.Render(w, r, ErrInvalidRequest(errors.New("no categories provided")))
		return
	}

	db ,err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// check that user exists
	var u sql.NullString
	err = db.QueryRow("SELECT login_name FROM Users WHERE login_name = ?;", login_name).Scan(&u)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("user not found")))
		return
	}

	// Check auth token
	var req_user_id, req_login_name string
	claims, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if len(claims) > 0 {
		req_user_id = claims["user_id"].(string)
		req_login_name = claims["login_name"].(string)
	}

	// Prepare SQL to get submitted / tagged / copied links from User
	// (Start with queries for signed-out user, append additional ones if needed)

	base_fields := `SELECT 
		Links.id as link_id, 
		url, 
		submitted_by as login_name, 
		submit_date, 
		categories, 
		coalesce(global_summary,"") as summary, 
		coalesce(summary_count,0) as summary_count, 
		coalesce(like_count,0) as like_count, 
		coalesce(img_url,"") as img_url
	`

	// Get submitted links, replacing global categories with user-assigned
	submitted_from := fmt.Sprintf(` FROM Links
		JOIN
		(
			SELECT categories, link_id as tag_link_id
			FROM Tags
			WHERE submitted_by = '%s'
		)
		ON link_id = tag_link_id
		LEFT JOIN
		(
			SELECT count(*) as like_count, link_id as like_link_id
			FROM 'Link Likes'
			GROUP BY link_id
		)
		ON like_link_id = link_id
		LEFT JOIN
		(
			SELECT count(*) as summary_count, link_id as summary_link_id
			FROM Summaries
			GROUP BY link_id
		)
		ON summary_link_id = link_id`, login_name)
	
	submitted_where := fmt.Sprintf(` WHERE submitted_by = '%s'`, login_name)

	// Append category filters to submitted_where
	cats_split := strings.Split(categories, ",")
	for _, cat := range cats_split {
		submitted_where += fmt.Sprintf(` AND ',' || categories || ',' LIKE '%%,%s,%%'`, cat)
	}

	// Get tagged links submitted by other users, replacing global categories with user-assigned
	tagged_from := submitted_from
	tagged_where := fmt.Sprintf(` WHERE submitted_by != '%s'`, login_name)

	// Append category filters to tagged_where
	for _, cat := range cats_split {
		tagged_where += fmt.Sprintf(` AND ',' || categories || ',' LIKE '%%,%s,%%'`, cat)
	}

	// Get copied links
	copied_fields := strings.Replace(base_fields, "categories", `coalesce(global_cats,"") as categories`, 1)
	copied_from := fmt.Sprintf(` FROM Links
	JOIN
		(
		SELECT link_id as copy_link_id, user_id as copier_id
		FROM 'Link Copies'
		JOIN Users
		ON Users.id = copier_id
		WHERE Users.login_name = '%s'
		)
	ON copy_link_id = link_id
	LEFT JOIN
		(
		SELECT count(*) as like_count, link_id as like_link_id
		FROM 'Link Likes'
		GROUP BY link_id
		)
	ON like_link_id = link_id
	LEFT JOIN
		(
		SELECT count(*) as summary_count, link_id as summary_link_id
		FROM Summaries
		GROUP BY link_id
		)
	ON summary_link_id = link_id`, login_name)

	copied_where := fmt.Sprintf(` WHERE ',' || categories || ',' LIKE '%%,%s,%%'`, cats_split[0])
	for _, cat := range cats_split[1:] {
		copied_where += fmt.Sprintf(` AND ',' || categories || ',' LIKE '%%,%s,%%'`, cat)
	}
	copied_where += fmt.Sprintf(` AND link_id NOT IN
	(
		SELECT link_id
		FROM TAGS
		WHERE submitted_by = '%s'
	);`, login_name)

	// Append additional queries for IsLiked, IsTagged, and IsCopied fields if auth claims verified
	if req_user_id != "" {
		added_fields := `, 
		coalesce(is_liked,0) as is_liked, 
		coalesce(is_tagged,0) as is_tagged,
		coalesce(is_copied,0) as is_copied`

		added_from := fmt.Sprintf(` LEFT JOIN
			(
			SELECT id, count(*) as is_liked, user_id, link_id as like_link_id2
			FROM 'Link Likes'
			WHERE user_id = '%[1]s'
			GROUP BY id
			)
		ON like_link_id2 = link_id 
		LEFT JOIN 
		(
			SELECT id as tag_id, link_id as tlink_id, count(*) as is_tagged 
			FROM Tags
			WHERE Tags.submitted_by = '%[2]s'
			GROUP BY tag_id
		)
		ON tlink_id = link_id
		LEFT JOIN
			(
			SELECT id as copy_id, count(*) as is_copied, user_id as cuser_id, link_id as clink_id
			FROM 'Link Copies'
			WHERE cuser_id = '%[1]s'
			GROUP BY copy_id
			)
		ON clink_id = link_id`, req_user_id, req_login_name)

		// Submitted
		submitted_fields := base_fields + added_fields
		submitted_from += added_from
		submitted_sql := submitted_fields + submitted_from + submitted_where

		// Tagged
		tagged_fields := base_fields + added_fields
		tagged_from += added_from
		tagged_sql := tagged_fields + tagged_from + tagged_where

		// Copied
		copied_fields += added_fields
		copied_from += added_from
		copied_sql := copied_fields + copied_from + copied_where

		// Scan links
		var submitted, tagged, copied *[]model.LinkSignedIn
		for _, sql := range []string{submitted_sql, tagged_sql, copied_sql} {
			rows, err := db.Query(sql)
			if err != nil {
				log.Fatal(err)
			}
			defer rows.Close()
			
			switch sql {
			case submitted_sql:
				submitted = ScanTmapLinksSignedIn(db, rows)
			case tagged_sql:
				tagged = ScanTmapLinksSignedIn(db, rows)
			case copied_sql:
				copied = ScanTmapLinksSignedIn(db, rows)
			}
		}

		// Add links to tmap
		tmap := model.TreasureMap[model.LinkSignedIn]{Submitted: submitted, Tagged: tagged, Copied: copied}

		// Get subcategory counts
		all_links := slices.Concat(*submitted, *tagged, *copied)
		cat_counts := GetTmapCategoryCounts(&all_links, cats_split)

		// Add subcategories to tmap
		tmap.Categories = cat_counts
		render.JSON(w, r, tmap)
		
	// User not signed in: omit isLiked / isCopied / isTagged fields
	} else {
		submitted_sql := base_fields + submitted_from + submitted_where
		tagged_sql := base_fields + tagged_from + tagged_where
		copied_sql := copied_fields + copied_from + copied_where

		// Scan links
		var submitted, tagged, copied *[]model.LinkSignedOut
		for _, sql := range []string{submitted_sql, tagged_sql, copied_sql} {
			rows, err := db.Query(sql)
			if err != nil {
				log.Fatal(err)
			}
			defer rows.Close()
			
			switch sql {
			case submitted_sql:
				submitted = ScanTmapLinksSignedOut(db, rows)
			case tagged_sql:
				tagged = ScanTmapLinksSignedOut(db, rows)
			case copied_sql:
				copied = ScanTmapLinksSignedOut(db, rows)
			}
		}

		// Add links to tmap
		tmap := model.TreasureMap[model.LinkSignedOut]{Submitted: submitted, Tagged: tagged, Copied: copied}

		// Get subcategory counts
		all_links := slices.Concat(*submitted, *tagged, *copied)
		cat_counts := GetTmapCategoryCounts(&all_links, cats_split)

		// Add subcategories to tmap
		tmap.Categories = cat_counts
		render.JSON(w, r, tmap)
	}
}

func ScanTmapLinksSignedIn (db *sql.DB, rows *sql.Rows) *[]model.LinkSignedIn {
	var links = []model.LinkSignedIn{}
	for rows.Next() {
		i := model.LinkSignedIn{}
		err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.Categories, &i.Summary, &i.SummaryCount, &i.LikeCount, &i.ImgURL, &i.IsLiked, &i.IsTagged, &i.IsCopied)
		if err != nil {
			log.Fatal(err)
		}
		links = append(links, i)
	}

	return &links
}

func ScanTmapLinksSignedOut (db *sql.DB, rows *sql.Rows) *[]model.LinkSignedOut {	
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

// Get counts of each category found in links
// Omit any categories passed via omitted_cats
// (omit used to retrieve subcategories by passing directly searched categories)
func GetTmapCategoryCounts[T model.Link] (links *[]T, omitted_cats []string) *[]model.CategoryCount {
	cat_counts := []model.CategoryCount{}
	cats_found := []string{}
	var cat_found bool

	// for each link in links
	for _, link := range *links {

		// for each category in categories (comma-separated string)
		for _, cat := range strings.Split(link.GetCategories(), ",") {

			// skip if category is in omitted_cats
			if omitted_cats != nil && slices.Contains(omitted_cats, cat) {
				continue
			}

			// check if category is already in cat_counts
			cat_found = false
			for _, fc := range cats_found {

				// if found
				if fc == cat {
					cat_found = true

					// find slice with category and increment count
					for i, count := range cat_counts {
						if count.Category == cat {
							cat_counts[i].Count++
							break
						}
					}
				}
			}

			// add new category to cat_counts with fresh count if not found
			if !cat_found {
				cat_counts = append(cat_counts, model.CategoryCount{Category: cat, Count: 1})

				// add to found categories
				cats_found = append(cats_found, cat)
			}
		}
	}

	// sort categories by count
	slices.SortFunc(cat_counts, model.SortCategories)

	// limit to top 5 categories for now
	CATEGORY_LIMIT := 5
	if len(cat_counts) > CATEGORY_LIMIT {
		cat_counts = cat_counts[:CATEGORY_LIMIT]
	}

	return &cat_counts
}