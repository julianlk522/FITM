package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"golang.org/x/crypto/bcrypt"

	"oitm/model"
)

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

	// TODO: generate and return jwt along with login name
	var token string = "token"
	return_json := map[string]string{"token": token, "login_name": signup_data.UserAuth.LoginName}

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

	// Attempt to collect user hashed password, 
	// Abort if user not found
	var p sql.NullString
	err = db.QueryRow("SELECT password FROM Users WHERE login_name = ?", login_data.LoginName).Scan(&p)
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

	// TODO: generate and return jwt along with login name
	var token string = "token"
	return_json := map[string]string{"token": token, "login_name": login_data.LoginName}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, return_json)
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
	
	return_json := map[string]string{"token": edit_profile_data.AuthToken}
	
	// TODO: check auth token

	// About
	if edit_profile_data.EditAboutRequest != nil {
		// TODO replace hard-coded id with id corresponding
		// to provided auth token
		_, err = db.Exec(`UPDATE Users SET about = ? WHERE id = ?`, edit_profile_data.EditAboutRequest.About, edit_profile_data.AuthToken)
		if err != nil {
			log.Fatal(err)
		}

		return_json["about"] = edit_profile_data.EditAboutRequest.About
	}
	
	// Profile Pic
	if edit_profile_data.EditPfpRequest != nil {
		// TODO replace hard-coded id with id corresponding
		// to provided auth token
		_, err = db.Exec(`UPDATE Users SET pfp = ? WHERE id = ?`, edit_profile_data.EditPfpRequest.PFP, edit_profile_data.AuthToken)
		if err != nil {
			log.Fatal(err)
		}

		return_json["pfp"] = edit_profile_data.EditPfpRequest.PFP
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, return_json)
}

// GET USER TREASURE MAP
// (includes links tagged by and copied by user)
func GetTreasureMap(w http.ResponseWriter, r *http.Request) {
	var user_id string = chi.URLParam(r, "id")

	db ,err := sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// get user-assigned tag categories
	get_custom_cats_sql := fmt.Sprintf(`SELECT link_id, categories as cats FROM Tags JOIN Users
	ON Users.login_name = Tags.submitted_by
	WHERE Users.id = '%s'`, user_id)

	rows, err := db.Query(get_custom_cats_sql)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	user_custom_cats := []model.CustomLinkCategories{}
	for rows.Next() {
		i := model.CustomLinkCategories{}
		err := rows.Scan(&i.LinkID, &i.Categories)
		if err != nil {
			log.Fatal(err)
		}
		user_custom_cats = append(user_custom_cats, i)
	}

	// get all map links and their global categories + like counts
	get_map_sql := fmt.Sprintf(`SELECT Links.id as link_id, url, submitted_by, submit_date, coalesce(global_cats,"") as global_cats, coalesce(like_count,0) as like_count FROM LINKS LEFT JOIN (SELECT link_id as like_link_id, count(*) as like_count FROM 'Link Likes' GROUP BY like_link_id) ON Links.id = like_link_id WHERE link_id IN ( SELECT link_id FROM Tags JOIN Users ON Users.login_name = Tags.submitted_by WHERE Users.id = '%s' UNION SELECT link_id FROM (SELECT link_id, NULL as cats, user_id as link_copier_id FROM 'Link Copies' JOIN Users ON Users.id = link_copier_id WHERE link_copier_id = '%s') JOIN Links ON Links.id = link_id) ORDER BY like_count DESC, link_id ASC;`, user_id, user_id)

	links := []model.Link{}
	rows, err = db.Query(get_map_sql)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		i := model.Link{}
		err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.Categories, &i.LikeCount)
		if err != nil {
			log.Fatal(err)
		}
		links = append(links, i)
	}

	// replace global categories for links to which the user has submitted their own tags
	for i, link := range links {
		for _, cat := range user_custom_cats {
			if link.ID == cat.LinkID {
				links[i].Categories = cat.Categories
			}
		}
	}

	render.JSON(w, r, links)
	render.Status(r, http.StatusOK)
}