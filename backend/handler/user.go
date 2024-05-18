package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/exp/slices"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"golang.org/x/crypto/bcrypt"

	"oitm/model"
)

var work_dir string
var pic_dir string
var pic_files http.FileSystem

func init() {
	work_dir, _ = os.Getwd()
	pic_dir = filepath.Join(work_dir, "db/profile-pics")
	pic_files = http.Dir(pic_dir)
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
	_, claims, err := jwtauth.FromContext(r.Context())
	// claims = {"user_id":"1234","login_name":"johndoe"}
	if err != nil {
		log.Fatal(err)
	}
	req_user_id, ok := claims["user_id"]
	if !ok {
		log.Fatal("invalid auth token")
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
	rctx := chi.RouteContext(r.Context())
	path_prefix := strings.TrimSuffix(rctx.RoutePattern(), "{file_name}")

	fs := http.StripPrefix(path_prefix, http.FileServer(pic_files))
	fs.ServeHTTP(w, r)
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

	// get user-assigned tag categories
	get_custom_cats_sql := fmt.Sprintf(`SELECT link_id, categories as cats FROM Tags JOIN Users
	ON Users.login_name = Tags.submitted_by
	WHERE Users.login_name = '%s'`, login_name)

	rows, err := db.Query(get_custom_cats_sql)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var user_custom_cats []model.CustomLinkCategories
	for rows.Next() {
		i := model.CustomLinkCategories{}
		err := rows.Scan(&i.LinkID, &i.Categories)
		if err != nil {
			log.Fatal(err)
		}
		user_custom_cats = append(user_custom_cats, i)
	
	}
	

	// get all map links and their global categories + like counts,
	get_map_sql := fmt.Sprintf(`SELECT Links.id as link_id, url, submitted_by, submit_date, coalesce(global_cats,"") as global_cats, coalesce(like_count,0) as like_count FROM LINKS LEFT JOIN (SELECT link_id as like_link_id, count(*) as like_count FROM 'Link Likes' GROUP BY like_link_id) ON Links.id = like_link_id WHERE link_id IN ( SELECT link_id FROM Tags JOIN Users ON Users.login_name = Tags.submitted_by WHERE Users.login_name = '%s' UNION SELECT link_id FROM (SELECT link_id, NULL as cats, user_id as link_copier_id FROM 'Link Copies' JOIN Users ON Users.id = link_copier_id WHERE Users.login_name = '%s') JOIN Links ON Links.id = link_id) ORDER BY like_count DESC, link_id ASC;`, login_name, login_name)
	rows, err = db.Query(get_map_sql)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	if !rows.Next() {
		render.Status(r, http.StatusOK)
		render.JSON(w, r, model.TreasureMap{})
	}
	
	links := []model.Link{}
	for rows.Next() {
		i := model.Link{}
		err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.Categories, &i.LikeCount)
		if err != nil {
			log.Fatal(err)
		}
		links = append(links, i)
	}

	// replace global categories for links to which the user has submitted their own tags
	if len(user_custom_cats) > 0 {
		for l, link := range links {
			for c, cat := range user_custom_cats {
				if link.ID == cat.LinkID {

					// replace scanned categories with user's
					links[l].Categories = cat.Categories

					// remove custom cats from slice to speed up remaining lookups
					user_custom_cats = append(user_custom_cats[:c], user_custom_cats[c+1:]...)
				}
			}
		}	
	}

	// get category counts
	cat_counts := []model.CategoryCount{}
	cats_found := []string{}
	var cat_found bool
	for _, link := range links {

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
	tmap := model.TreasureMap{Links: links, Categories: cat_counts}

	render.JSON(w, r, tmap)
	render.Status(r, http.StatusOK)
}