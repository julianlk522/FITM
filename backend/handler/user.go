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
	var id int64
	err = db.QueryRow("SELECT id FROM Users WHERE login_name = ?", signup_data.UserAuth.LoginName).Scan(&id)
	if err != nil {
		log.Fatal(err)
	}

	// generate and return jwt containing user ID and login_name
	token_data := map[string]interface{}{"user_id": id, "login_name": signup_data.LoginName}
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
	var req_user_id string
	claims, err := GetJWTClaims(r)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if len(claims) > 0 {
		req_user_id = claims["user_id"].(string)
	}

	// Get links
	// User signed in: get isLiked for each link
	if req_user_id != "" {

		// Get submitted links, replacing global categories with user-assigned
		get_submitted_sql := fmt.Sprintf(`SELECT Links.id as link_id, url, submitted_by as login_name, submit_date, categories, coalesce(global_summary,"") as summary, coalesce(summary_count,0) as summary_count, coalesce(like_count,0) as like_count, coalesce(is_liked,0) as is_liked
		FROM Links
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
					SELECT id, count(*) as is_liked, user_id, link_id as like_link_id2
					FROM 'Link Likes'
					WHERE user_id == %s
					GROUP BY id
					)
			ON like_link_id2 = link_id
			LEFT JOIN
				(
				SELECT count(*) as summary_count, link_id as summary_link_id
				FROM Summaries
				GROUP BY link_id
				)
			ON summary_link_id = link_id
		WHERE login_name = '%s';`, login_name, req_user_id, login_name)
		rows, err := db.Query(get_submitted_sql)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		// Scan submitted
		submitted := []model.LinkSignedIn{}
		for rows.Next() {
			i := model.LinkSignedIn{}
			err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.Categories, &i.Summary, &i.SummaryCount, &i.LikeCount, &i.IsLiked)
			if err != nil {
				log.Fatal(err)
			}
			submitted = append(submitted, i)
		}

		// Get tagged links submitted by other users
		get_tagged_sql := fmt.Sprintf(`SELECT Links.id as link_id, url, submitted_by as login_name, submit_date, categories, coalesce(global_summary,"") as summary, coalesce(summary_count,0) as summary_count, coalesce(like_count,0) as like_count, coalesce(is_liked,0) as is_liked
		FROM Links
		JOIN
			(
			SELECT categories, link_id as tag_link_id
			FROM Tags
			WHERE submitted_by = '%s'
			)
		ON tag_link_id = link_id
	LEFT JOIN
			(
			SELECT count(*) as like_count, link_id as like_link_id
			FROM 'Link Likes'
			GROUP BY link_id
			)
	ON like_link_id = link_id
	LEFT JOIN
			(
			SELECT id, count(*) as is_liked, user_id, link_id as like_link_id2
			FROM 'Link Likes'
			WHERE user_id == %s
			GROUP BY id
			)
	ON like_link_id2 = link_id
	LEFT JOIN
		(
		SELECT count(*) as summary_count, link_id as summary_link_id
		FROM Summaries
		GROUP BY link_id
		)
	ON summary_link_id = link_id
	WHERE login_name != '%s';`, login_name, req_user_id, login_name)
		rows, err = db.Query(get_tagged_sql)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		// Scan tagged
		tagged := []model.LinkSignedIn{}
		for rows.Next() {
			i := model.LinkSignedIn{}
			err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.Categories, &i.Summary, &i.SummaryCount, &i.LikeCount, &i.IsLiked)
			if err != nil {
				log.Fatal(err)
			}
			tagged = append(tagged, i)
		}

		// Get copied links
		get_copied_sql := fmt.Sprintf(`SELECT Links.id as link_id, url, submitted_by as login_name, submit_date, coalesce(global_cats,"") as categories, coalesce(global_summary,"") as summary, coalesce(summary_count,0) as summary_count, coalesce(like_count,0) as like_count, coalesce(is_liked,0) as is_liked
		FROM Links
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
			SELECT id, count(*) as is_liked, user_id, link_id as like_link_id2
			FROM 'Link Likes'
			WHERE user_id == %s
			GROUP BY id
			)
		ON like_link_id2 = link_id
		LEFT JOIN
			(
			SELECT count(*) as summary_count, link_id as summary_link_id
			FROM Summaries
			GROUP BY link_id
			)
		ON summary_link_id = link_id;`, login_name, req_user_id)
		rows, err = db.Query(get_copied_sql)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		// Scan copied
		copied := []model.LinkSignedIn{}
		for rows.Next() {
			i := model.LinkSignedIn{}
			err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.Categories, &i.Summary, &i.SummaryCount, &i.LikeCount, &i.IsLiked)
			if err != nil {
				log.Fatal(err)
			}
			copied = append(copied, i)
		}

		// Add links to tmap
		tmap := model.TreasureMapSignedIn{Submitted: submitted, Tagged: tagged, Copied: copied}

		// get category counts
		cat_counts := []model.CategoryCount{}
		cats_found := []string{}
		var cat_found bool
		for _, link := range slices.Concat(submitted, tagged, copied) {

			// for each category in the comma-separated string,
			for _, cat := range strings.Split(link.Categories, ",") {

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

				// else add to slice with fresh count
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
		
		// combine links and categories in response
		tmap.Categories = cat_counts
		render.JSON(w, r, tmap)
		
	// User not signed in: omit isLiked
	} else {

		// Get submitted links, replacing global categories with user-assigned
		get_submitted_sql := fmt.Sprintf(`SELECT Links.id as link_id, url, submitted_by as login_name, submit_date, categories, coalesce(global_summary,"") as summary, coalesce(summary_count,0) as summary_count, coalesce(like_count,0) as like_count
		FROM Links
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
		ON summary_link_id = link_id
		WHERE login_name = '%s';`, login_name, login_name)
		rows, err := db.Query(get_submitted_sql)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		// Scan submitted
		submitted := []model.Link{}
		for rows.Next() {
			i := model.Link{}
			err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.Categories, &i.Summary, &i.SummaryCount, &i.LikeCount)
			if err != nil {
				log.Fatal(err)
			}
			submitted = append(submitted, i)
		}

		// Get tagged links submitted by other users
		get_tagged_sql := fmt.Sprintf(`SELECT Links.id as link_id, url, submitted_by as login_name, submit_date, categories, coalesce(global_summary,"") as summary, coalesce(summary_count,0) as summary_count, coalesce(like_count,0) as like_count
		FROM Links
		JOIN
			(
			SELECT categories, link_id as tag_link_id
			FROM Tags
			WHERE submitted_by = '%s'
			)
		ON tag_link_id = link_id
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
		ON summary_link_id = link_id
	WHERE login_name != '%s';`, login_name, login_name)
		rows, err = db.Query(get_tagged_sql)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		// Scan tagged
		tagged := []model.Link{}
		for rows.Next() {
			i := model.Link{}
			err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.Categories, &i.Summary, &i.SummaryCount, &i.LikeCount)
			if err != nil {
				log.Fatal(err)
			}
			tagged = append(tagged, i)
		}

		// Get copied links
		get_copied_sql := fmt.Sprintf(`SELECT Links.id as link_id, url, submitted_by as login_name, submit_date, coalesce(global_cats,"") as categories, coalesce(global_summary,"") as summary, coalesce(summary_count,0) as summary_count, coalesce(like_count,0) as like_count
		FROM Links
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
		ON summary_link_id = link_id;`, login_name)
		rows, err = db.Query(get_copied_sql)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		// Scan copied
		copied := []model.Link{}
		for rows.Next() {
			i := model.Link{}
			err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.Categories, &i.Summary, &i.SummaryCount, &i.LikeCount)
			if err != nil {
				log.Fatal(err)
			}
			copied = append(copied, i)
		}

		// Add links to tmap
		tmap := model.TreasureMap{Submitted: submitted, Tagged: tagged, Copied: copied}

		// get category counts
		cat_counts := []model.CategoryCount{}
		cats_found := []string{}
		var cat_found bool
		for _, link := range slices.Concat(submitted, tagged, copied) {

			// for each category in the comma-separated string,
			for _, cat := range strings.Split(link.Categories, ",") {

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

				// else add to slice with fresh count
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
		
		// combine links and categories in response
		tmap.Categories = cat_counts
		render.JSON(w, r, tmap)
	}
}