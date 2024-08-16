package handler

import (
	"errors"
	"image"
	"io"
	"log"
	"net/http"
	util "oitm/handler/util"
	"os"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/webp"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"oitm/db"
	e "oitm/error"
	m "oitm/middleware"
	"oitm/model"
)

var pic_dir string

func init() {
	work_dir, _ := os.Getwd()
	pic_dir = filepath.Join(work_dir, "db/profile-pics")
}

// Auth
func SignUp(w http.ResponseWriter, r *http.Request) {
	signup_data := &model.SignUpRequest{}

	if err := render.Bind(r, signup_data); err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	if util.LoginNameTaken(signup_data.Auth.LoginName) {
		render.Render(w, r, e.ErrInvalidRequest(errors.New("login name taken")))
		return
	}

	pw_hash, err := bcrypt.GenerateFromPassword([]byte(signup_data.Auth.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Client.Exec(`INSERT INTO users VALUES (?,?,?,?,?,?)`, nil, signup_data.Auth.LoginName, pw_hash, nil, nil, signup_data.CreatedAt)
	if err != nil {
		log.Fatal(err)
	}

	token, err := util.GetJWTFromLoginName(signup_data.Auth.LoginName)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	util.RenderJWT(token, w, r)
}

func LogIn(w http.ResponseWriter, r *http.Request) {
	login_data := &model.LogInRequest{}

	if err := render.Bind(r, login_data); err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	is_authenticated, err := util.AuthenticateUser(login_data.LoginName, login_data.Password)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	} else if !is_authenticated {
		render.Render(w, r, e.ErrInvalidRequest(errors.New("invalid login")))
		return
	}

	token, err := util.GetJWTFromLoginName(login_data.Auth.LoginName)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	util.RenderJWT(token, w, r)
}

// Treasure map
func EditAbout(w http.ResponseWriter, r *http.Request) {
	edit_about_data := &model.EditAboutRequest{}
	if err := render.Bind(r, edit_about_data); err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}

	req_user_id := r.Context().Value(m.UserIDKey).(string)
	_, err := db.Client.Exec(`UPDATE Users SET about = ? WHERE id = ?`, edit_about_data.About, req_user_id)
	if err != nil {
		log.Fatal(err)
	}
	
	render.Status(r, http.StatusOK)
	render.JSON(w, r, edit_about_data)
}

func GetProfilePic(w http.ResponseWriter, r *http.Request) {
	// (from backend/db/profile-pics/{file_name})

	var file_name string = chi.URLParam(r, "file_name")
	path := pic_dir + "/" + file_name
	
	if _, err := os.Stat(path); err != nil {
		render.Render(w, r, e.ErrInvalidRequest(errors.New("profile pic not found")))
		return
	}
	
	http.ServeFile(w, r, path)
}

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
	if !util.HasAcceptableAspectRatio(img) {
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
	_, err = db.Client.Exec(`UPDATE Users SET pfp = ? WHERE id = ?`, unique_name, req_user_id)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(errors.New("could not save new profile pic")))
		return
	}

	http.ServeFile(w, r, full_path)
}

func GetTreasureMap(w http.ResponseWriter, r *http.Request) {
	var login_name string = chi.URLParam(r, "login_name")
	if login_name == "" {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoLoginName))
		return
	}

	user_exists, err := util.UserExists(login_name)
	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	} else if !user_exists {
		render.Render(w, r, e.ErrInvalidRequest(e.ErrNoUserWithLoginName))
		return
	}

	var tmap interface{}

	req_user_id := r.Context().Value(m.UserIDKey).(string)
	if req_user_id != "" {
		tmap, err = util.GetTmapForUser[model.TmapLinkSignedIn](login_name, r)
	} else {
		tmap, err = util.GetTmapForUser[model.TmapLink](login_name, r)
	}

	if err != nil {
		render.Render(w, r, e.ErrInvalidRequest(err))
		return
	}
	
	render.JSON(w, r, tmap)
}