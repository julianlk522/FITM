package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"oitm/db"

	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
	"github.com/lestrrat-go/jwx/v2/jwt"
	_ "golang.org/x/image/webp"

	"time"

	"golang.org/x/crypto/bcrypt"
)

// Auth
func LoginNameTaken(login_name string) bool {
	var s sql.NullString
	if err := db.Client.QueryRow("SELECT login_name FROM Users WHERE login_name = ?", login_name).Scan(&s); err == nil {
		return true
	}
	return false
}

func AuthenticateUser(login_name string, password string) (bool, error) {
	var id, p sql.NullString
	if err := db.Client.QueryRow("SELECT id, password FROM Users WHERE login_name = ?", login_name).Scan(&id, &p); err != nil {
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

func GetJWTFromLoginName(login_name string) (string, error) {
	var id sql.NullString
	err := db.Client.QueryRow("SELECT id FROM Users WHERE login_name = ?", login_name).Scan(&id)
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

func RenderJWT(token string, w http.ResponseWriter, r *http.Request) {
	return_json := map[string]string{"token": token}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, return_json)
}



// Upload profile pic
func HasAcceptableAspectRatio(img image.Image) bool {
	b := img.Bounds()
	width, height := b.Max.X, b.Max.Y
	ratio := float64(width) / float64(height)

	if ratio > 2.0 || ratio < 0.5 {
		return false
	}

	return true
}