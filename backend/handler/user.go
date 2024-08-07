package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	_ "golang.org/x/image/webp"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"golang.org/x/crypto/bcrypt"

	query "oitm/db/query"
	e "oitm/error"
	m "oitm/middleware"
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
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	if _LoginNameTaken(signup_data.Auth.LoginName) {
		render.Render(w, r, e.ErrInvalidRequest(errors.New("login name taken")))
		return
	}

	pw_hash, err := bcrypt.GenerateFromPassword([]byte(signup_data.Auth.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	_, err = DBClient.Exec(`INSERT INTO users VALUES (?,?,?,?,?,?)`, nil, signup_data.Auth.LoginName, pw_hash, nil, nil, signup_data.CreatedAt)
	if err != nil {
		log.Fatal(err)
	}

	token, err := _GetJWTFromLoginName(signup_data.Auth.LoginName)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	_RenderJWT(token, w, r)
}

func _LoginNameTaken(login_name string) bool {
	var s sql.NullString
	if err := DBClient.QueryRow("SELECT login_name FROM Users WHERE login_name = ?", login_name).Scan(&s); err == nil {
		return true
	}
	return false
}

// LOG IN
func LogIn(w http.ResponseWriter, r *http.Request) {
	login_data := &model.LogInRequest{}

	if err := render.Bind(r, login_data); err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	is_authenticated, err := _AuthenticateUser(login_data.LoginName, login_data.Password)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	} else if !is_authenticated {
		render.Render(w, r, e.ErrInvalidRequest(errors.New("invalid login")))
		return
	}

	token, err := _GetJWTFromLoginName(login_data.Auth.LoginName)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	_RenderJWT(token, w, r)
}

func _AuthenticateUser(login_name string, password string) (bool, error) {
	var id, p sql.NullString
	if err := DBClient.QueryRow("SELECT id, password FROM Users WHERE login_name = ?", login_name).Scan(&id, &p); err != nil {
		if err == sql.ErrNoRows {
			return false, errors.New("user not found")
		}
		return false, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(p.String), []byte(password)); err != nil {
		return false, errors.New("incorrect password")
	}

	return true, nil
}

func _GetJWTFromLoginName(login_name string) (string, error) {
	var id sql.NullString
	err := DBClient.QueryRow("SELECT id FROM Users WHERE login_name = ?", login_name).Scan(&id)
	if err != nil {
		return "", err
	}

	claims := map[string]interface{}{"user_id": id.String, "login_name": login_name}

	// TODO: change jwt secret
	auth := jwtauth.New("HS256", []byte("secret"), nil, jwt.WithAcceptableSkew(24*time.Hour))
	_, token, err := auth.Encode(claims)
	if err != nil {
		return "", err
	}

	return token, nil
}

func _RenderJWT(token string, w http.ResponseWriter, r *http.Request) {
	return_json := map[string]string{"token": token}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, return_json)
}

func EditAbout(w http.ResponseWriter, r *http.Request) {
	edit_about_data := &model.EditAboutRequest{}
	if err := render.Bind(r, edit_about_data); err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	req_user_id := r.Context().Value(m.UserIDKey).(string)
	_, err := DBClient.Exec(`UPDATE Users SET about = ? WHERE id = ?`, edit_about_data.About, req_user_id)
	if err != nil {
		log.Fatal(err)
	}
	
	render.Status(r, http.StatusOK)
	render.JSON(w, r, edit_about_data)
}

// GET PROFILE PICTURE
// (from backend/db/profile-pics/{file_name})
func GetProfilePic(w http.ResponseWriter, r *http.Request) {
	var file_name string = chi.URLParam(r, "file_name")
	path := pic_dir + "/" + file_name
	
	if _, err := os.Stat(path); err != nil {
		render.Render(w, r, e.ErrInvalidRequest(errors.New("profile pic not found")))
		return
	}
	
	// Serve if found
	http.ServeFile(w, r, path)
}

// UPLOAD NEW PROFILE PICTURE
func UploadNewProfilePic(w http.ResponseWriter, r *http.Request) {

	// Get file (up to 10MB)
	r.ParseMultipartForm( 10 << 20 )
	file, handler, err := r.FormFile("pic")
    if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
        return
    }
    defer file.Close()

	// Valid image
	if !strings.Contains(handler.Header.Get("Content-Type"), "image") {
		render.Render(w, r, e.ErrInvalidRequest(errors.New("invalid file provided (accepted image formats: .jpg, .jpeg, .png, .webp)")))
		return
	}

	img, _, err := image.Decode(file)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}
	
	// Aspect ratio is no more than 2:1 and no less than 0.5:1
	if !_HasAcceptableAspectRatio(img) {
		render.Render(w, r, e.ErrInvalidRequest(errors.New("profile pic aspect ratio must be no more than 2:1 and no less than 0.5:1")))
		return
	}

	extension := filepath.Ext(handler.Filename)
	unique_name := uuid.New().String() + extension
	full_path := pic_dir + "/" + unique_name

	dst, err := os.Create(full_path)
	if err != nil {
		// Note: if, for some reason, the directory at pic_dir's path
		// doesn't exist, this will fail
		// shouldn't matter but just for posterity
		render.Render(w, r, e.ErrInvalidRequest(errors.New("could not create new file")))
		return
	}
	defer dst.Close()

	// Restore img file cursor to start
	file.Seek(0, 0)
	
	// Save to new file
	if _, err := io.Copy(dst, file); err != nil {
		render.Render(w, r, e.ErrInvalidRequest(errors.New("could not copy profile pic to new file")))
		return
	}

	req_user_id := r.Context().Value(m.UserIDKey).(string)
	_, err = DBClient.Exec(`UPDATE Users SET pfp = ? WHERE id = ?`, unique_name, req_user_id)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(errors.New("could not save new profile pic")))
		return
	}

	http.ServeFile(w, r, full_path)
}

func _HasAcceptableAspectRatio(img image.Image) bool {
	b := img.Bounds()
	width, height := b.Max.X, b.Max.Y
	ratio := float64(width) / float64(height)

	if ratio > 2.0 || ratio < 0.5 {
		return false
	}

	return true
}

// GET TREASURE MAP
func GetTreasureMap(w http.ResponseWriter, r *http.Request) {
	var login_name string = chi.URLParam(r, "login_name")
	if login_name == "" {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoLoginName))
		return
	}

	user_exists, err := _UserExists(login_name)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	} else if !user_exists {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoUserWithLoginName))
		return
	}

	req_user_id := r.Context().Value(m.UserIDKey).(string)
	if req_user_id != "" {
		_RenderTmap[model.TmapLinkSignedIn](r, w, login_name)
	} else {
		_RenderTmap[model.TmapLink](r, w, login_name)
	}	
}

func _UserExists(login_name string) (bool, error) {
	var u sql.NullString
	err := DBClient.QueryRow("SELECT id FROM Users WHERE login_name = ?;", login_name).Scan(&u)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

func _RenderTmap[T model.TmapLink | model.TmapLinkSignedIn](r *http.Request, w http.ResponseWriter, login_name string) {
	submitted_sql := query.NewTmapSubmitted(login_name)
	copied_sql := query.NewTmapCopied(login_name)
	tagged_sql := query.NewTmapTagged(login_name)
	
	cats_params := r.URL.Query().Get("cats") 
	has_cat_filter := cats_params != ""

	var cats []string
	var profile *model.Profile
	if has_cat_filter {
		cats = strings.Split(cats_params, ",")
	} else {
		var err error
		profile_sql := query.NewTmapProfile(login_name)
		profile, err = _ScanTmapProfile(profile_sql)
		if err != nil {
			render.Render(w, r, e.ErrInvalidRequest(err))
		}
	}

	if has_cat_filter {
		submitted_sql = submitted_sql.FromCategories(cats)
		copied_sql = copied_sql.FromCategories(cats)
		tagged_sql = tagged_sql.FromCategories(cats)
	}

	req_user_id := r.Context().Value(m.UserIDKey).(string)
	req_login_name := r.Context().Value(m.LoginNameKey).(string)

	// Requesting user signed in: get IsLiked / IsCopied / IsTagged for each link
	if req_user_id != "" {	
		submitted_sql = submitted_sql.AsSignedInUser(req_user_id, req_login_name)
		copied_sql = copied_sql.AsSignedInUser(req_user_id, req_login_name)
		tagged_sql = tagged_sql.AsSignedInUser(req_user_id, req_login_name)
	}

	submitted, err := _ScanTmapLinks[T](submitted_sql.Query)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
	}
	tagged, err := _ScanTmapLinks[T](tagged_sql.Query)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
	}
	copied, err := _ScanTmapLinks[T](copied_sql.Query)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
	}

	all_links := slices.Concat(*submitted, *tagged, *copied)
	var cat_counts *[]model.CategoryCount
	if has_cat_filter {
		cat_counts = GetTmapCategoryCounts(&all_links, cats)
	} else {
		cat_counts = GetTmapCategoryCounts(&all_links, nil)
	}

	sections := &model.TreasureMapSections[T]{
		Submitted: submitted,
		Tagged: tagged,
		Copied: copied,
		Categories: cat_counts,
	}

	if has_cat_filter {
		tmap := model.FilteredTreasureMap[T]{
			TreasureMapSections: sections,
		}
		render.JSON(w, r, tmap)

	} else {
		tmap := model.TreasureMap[T]{
			Profile: profile, 
			TreasureMapSections: sections,
		}
		render.JSON(w, r, tmap)
	}
}

func _ScanTmapProfile(sql string) (*model.Profile, error) {
	var u model.Profile
	err := DBClient.QueryRow(sql).Scan(&u.LoginName, &u.About, &u.PFP, &u.Created)
	if err != nil {
		return nil, e.ErrNoUserWithLoginName
	}

	return &u, nil
}

func _ScanTmapLinks[T model.TmapLink | model.TmapLinkSignedIn](sql query.Query) (*[]T, error) {
	rows, err := DBClient.Query(sql.Text)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links interface{}

	switch any(new(T)).(type) {
		case *model.TmapLinkSignedIn:
			var signed_in_links = []model.TmapLinkSignedIn{}

			for rows.Next() {
				i := model.TmapLinkSignedIn{}
				err := rows.Scan(
					&i.ID, 
					&i.URL, 
					&i.SubmittedBy, 
					&i.SubmitDate, 
					&i.Categories, 
					&i.CategoriesFromUser, 
					&i.Summary, 
					&i.SummaryCount, 
					&i.LikeCount, 
					&i.ImgURL,
					
					// Add IsLiked / IsCopied / IsTagged 
					&i.IsLiked, 
					&i.IsTagged, 
					&i.IsCopied)
				if err != nil {
					return nil, err
				}
				signed_in_links = append(signed_in_links, i)
			}

			links = &signed_in_links

		case *model.TmapLink:
			var signed_out_links = []model.TmapLink{}

			for rows.Next() {
				i := model.TmapLink{}
				err := rows.Scan(
					&i.ID, 
					&i.URL, 
					&i.SubmittedBy, 
					&i.SubmitDate, 
					&i.Categories, 
					&i.CategoriesFromUser, 
					&i.Summary, 
					&i.SummaryCount, 
					&i.LikeCount, 
					&i.ImgURL)
				if err != nil {
					return nil, err
				}
				signed_out_links = append(signed_out_links, i)
			}

			links = &signed_out_links
	}

	return links.(*[]T), nil	
}

// Get counts of each category found in links
// Omit any categories passed via omitted_cats
// (omit used to retrieve subcategories by passing directly searched categories)
// TODO: refactor to make this clearer
func GetTmapCategoryCounts[T model.TmapLink | model.TmapLinkSignedIn] (links *[]T, omitted_cats []string) *[]model.CategoryCount {
	counts := []model.CategoryCount{}
	found_cats := []string{}
	var found bool

	for _, link := range *links {
		var categories string
		switch l := any(link).(type) {
			case model.TmapLinkSignedIn:
				categories = l.Categories
			case model.TmapLink:
				categories = l.Categories
		}

		for _, cat := range strings.Split(categories, ",") {
			if omitted_cats != nil && slices.Contains(omitted_cats, cat) {
				continue
			}

			found = false
			for _, found_cat := range found_cats {
				if cat == found_cat {
					found = true

					for i, count := range counts {
						if count.Category == cat {
							counts[i].Count++
							break
						}
					}
				}
			}

			if !found {
				counts = append(counts, model.CategoryCount{Category: cat, Count: 1})

				// add to found categories
				found_cats = append(found_cats, cat)
			}
		}
	}

	_SortAndLimitTmapCategoryCounts(&counts)

	return &counts
}

func _SortAndLimitTmapCategoryCounts(counts *[]model.CategoryCount) {
	slices.SortFunc(*counts, model.SortCategories)

	if len(*counts) > TMAP_CATEGORY_COUNT_LIMIT {
		*counts = (*counts)[:TMAP_CATEGORY_COUNT_LIMIT]
	}
}