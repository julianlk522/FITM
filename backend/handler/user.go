package handler

import (
	"database/sql"
	"errors"
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

	if _LoginNameTaken(signup_data.UserAuth.LoginName) {
		render.Render(w, r, e.ErrInvalidRequest(errors.New("login name taken")))
		return
	}

	pw_hash, err := bcrypt.GenerateFromPassword([]byte(signup_data.UserAuth.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	_, err = DBClient.Exec(`INSERT INTO users VALUES (?,?,?,?,?,?)`, nil, signup_data.UserAuth.LoginName, pw_hash, nil, nil, signup_data.Created)
	if err != nil {
		log.Fatal(err)
	}

	token, err := _GetJWTFromLoginName(signup_data.UserAuth.LoginName)
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

	token, err := _GetJWTFromLoginName(login_data.UserAuth.LoginName)
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

// GET USER TREASURE MAP
// (includes tagged / copied links and category sum counts)
// (and all links submitted by user, since submission requires tag)
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
	// Requesting user signed in: get IsLiked / IsCopied / IsTagged for each link
	if req_user_id != "" {	
		tmap, err := _BuildTmap[model.TmapLinkSignedIn](login_name, r)
		if err != nil {
			render.Render(w, r, e.ErrInvalidRequest(err))
			return
		}
		render.JSON(w, r, tmap)
		
	// No auth
	} else {
		tmap, err := _BuildTmap[model.TmapLinkSignedOut](login_name, r)
		if err != nil {
			render.Render(w, r, e.ErrInvalidRequest(err))
			return
		}
		render.JSON(w, r, tmap)
	}
}

func GetTreasureMapByCategories(w http.ResponseWriter, r *http.Request) {
	var login_name string = chi.URLParam(r, "login_name")
	if login_name == "" {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoLoginName))
		return
	}

	var categories string = chi.URLParam(r, "categories")
	if categories == "" {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoCategories))
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
	
	split_cats := strings.Split(categories, ",")

	req_user_id := r.Context().Value(m.UserIDKey).(string)
	if req_user_id != "" {
		tmap, err := _BuildTmapFromCategories[model.TmapLinkSignedIn](login_name, split_cats, r)
		if err != nil {
			render.Render(w, r, e.ErrInvalidRequest(err))
			return
		}
		render.JSON(w, r, tmap)
	} else {
		tmap, err := _BuildTmapFromCategories[model.TmapLinkSignedOut](login_name, split_cats, r)
		if err != nil {
			render.Render(w, r, e.ErrInvalidRequest(err))
			return
		}
		render.JSON(w, r, tmap)
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

func _BuildTmap[T model.TmapLinkSignedIn | model.TmapLinkSignedOut](login_name string, r *http.Request) (*model.TreasureMap[T], error) {
	var submitted_sql *query.GetTmapSubmitted
	var copied_sql *query.GetTmapCopied
	var tagged_sql *query.GetTmapTagged

	req_user_id := r.Context().Value(m.UserIDKey).(string)
	req_login_name := r.Context().Value(m.LoginNameKey).(string)
	
	// Requesting user signed in: get IsLiked / IsCopied / IsTagged for each link
	if req_user_id != "" {	
		submitted_sql = query.NewGetTmapSubmitted(req_user_id, req_login_name).ForUser(login_name)
		copied_sql = query.NewGetTmapCopied(req_user_id, req_login_name).ForUser(login_name)
		tagged_sql = query.NewGetTmapTagged(req_user_id, req_login_name).ForUser(login_name)
		
	// No auth
	} else {
		submitted_sql = query.NewGetTmapSubmitted("", "").ForUser(login_name)
		copied_sql = query.NewGetTmapCopied("", "").ForUser(login_name)
		tagged_sql = query.NewGetTmapTagged("", "").ForUser(login_name)
	}

	var u model.Profile
	err := DBClient.QueryRow(`
	SELECT login_name, COALESCE(about,"") as about, COALESCE(pfp,"") as pfp, COALESCE(created,"") as created 
	FROM Users 
	WHERE login_name = ?;`, login_name).Scan(&u.LoginName, &u.About, &u.PFP, &u.Created)
	if err != nil {
		return nil, e.ErrNoUserWithLoginName
	}

	submitted, err := _ScanTmapLinks[T](submitted_sql.Query)
	if err != nil {
		return nil, err
	}
	tagged, err := _ScanTmapLinks[T](tagged_sql.Query)
	if err != nil {
		return nil, err
	}
	copied, err := _ScanTmapLinks[T](copied_sql.Query)
	if err != nil {
		return nil, err
	}

	tmap := model.TreasureMap[T]{Profile: u, Submitted: submitted, Tagged: tagged, Copied: copied}

	all_links := slices.Concat(*submitted, *tagged, *copied)
	cat_counts := GetTmapCategoryCounts(&all_links, nil)
	tmap.Categories = cat_counts

	return &tmap, nil
}

func _BuildTmapFromCategories[T model.TmapLinkSignedIn | model.TmapLinkSignedOut](login_name string, categories []string, r *http.Request) (*model.FilteredTreasureMap[T], error) {
	req_user_id := r.Context().Value(m.UserIDKey).(string)
	req_login_name := r.Context().Value(m.LoginNameKey).(string)

	var submitted_sql *query.GetTmapSubmitted
	var copied_sql *query.GetTmapCopied
	var tagged_sql *query.GetTmapTagged

	// Requesting user signed in: get IsLiked / IsCopied / IsTagged for each link
	if req_user_id != "" {	
		submitted_sql = query.NewGetTmapSubmitted(req_user_id, req_login_name).FromCategories(categories).ForUser(login_name)
		copied_sql = query.NewGetTmapCopied(req_user_id, req_login_name).FromCategories(categories).ForUser(login_name)
		tagged_sql = query.NewGetTmapTagged(req_user_id, req_login_name).FromCategories(categories).ForUser(login_name)
		
	// No auth
	} else {
		submitted_sql = query.NewGetTmapSubmitted("", "").FromCategories(categories).ForUser(login_name)
		copied_sql = query.NewGetTmapCopied("", "").FromCategories(categories).ForUser(login_name)
		tagged_sql = query.NewGetTmapTagged("", "").FromCategories(categories).ForUser(login_name)
	}

	submitted, err := _ScanTmapLinks[T](submitted_sql.Query)
	if err != nil {
		return nil, err
	}
	tagged, err := _ScanTmapLinks[T](tagged_sql.Query)
	if err != nil {
		return nil, err
	}
	copied, err := _ScanTmapLinks[T](copied_sql.Query)
	if err != nil {
		return nil, err
	}
	tmap := model.FilteredTreasureMap[T]{Submitted: submitted, Tagged: tagged, Copied: copied}

	all_links := slices.Concat(*submitted, *tagged, *copied)
	cat_counts := GetTmapCategoryCounts(&all_links, categories)
	tmap.Categories = cat_counts

	return &tmap, nil
}

func _ScanTmapLinks[T model.TmapLinkSignedIn | model.TmapLinkSignedOut](sql query.Query) (*[]T, error) {
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
				err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.Categories, &i.CategoriesFromUser, &i.Summary, &i.SummaryCount, &i.LikeCount, &i.ImgURL, &i.IsLiked, &i.IsTagged, &i.IsCopied)
				if err != nil {
					return nil, err
				}
				signed_in_links = append(signed_in_links, i)
			}

			links = &signed_in_links

		case *model.TmapLinkSignedOut:
			var signed_out_links = []model.TmapLinkSignedOut{}

			for rows.Next() {
				i := model.TmapLinkSignedOut{}
				err := rows.Scan(&i.ID, &i.URL, &i.SubmittedBy, &i.SubmitDate, &i.Categories, &i.CategoriesFromUser, &i.Summary, &i.SummaryCount, &i.LikeCount, &i.ImgURL)
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
func GetTmapCategoryCounts[T model.TmapLinkSignedIn | model.TmapLinkSignedOut] (links *[]T, omitted_cats []string) *[]model.CategoryCount {
	counts := []model.CategoryCount{}
	found_cats := []string{}
	var found bool

	for _, link := range *links {
		var categories string
		switch l := any(link).(type) {
			case model.TmapLinkSignedIn:
				categories = l.Categories
			case model.TmapLinkSignedOut:
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