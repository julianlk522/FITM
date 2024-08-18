package handler

import (
	"image"
	"path/filepath"
	"runtime"

	_ "golang.org/x/image/webp"

	"os"
	"testing"
)

func TestLoginNameTaken(t *testing.T) {
	var test_login_names = []struct {
		login_name string
		Taken      bool
	}{
		{"akjlhsdflkjhhasdf", false},
		{"janedoe", false},
		{"goolian", true},
	}

	for _, l := range test_login_names {
		return_true := LoginNameTaken(l.login_name)
		if l.Taken && !return_true {
			t.Fatalf("expected login name %s to be taken", l.login_name)
		} else if !l.Taken && return_true {
			t.Fatalf("login name %s NOT taken, expected error", l.login_name)
		}
	}
}

func TestAuthenticateUser(t *testing.T) {
	var test_logins = []struct {
		LoginName          string
		Password           string
		ShouldAuthenticate bool
	}{
		{"goolian", "password", false},
		{"monkey", "monkey", true},
		{"monkey", "bananas", false},
	}

	for _, l := range test_logins {
		return_true, err := AuthenticateUser(l.LoginName, l.Password)
		if l.ShouldAuthenticate && !return_true {
			t.Fatalf("expected login name %s to be authenticated", l.LoginName)
		} else if !l.ShouldAuthenticate && return_true {
			t.Fatalf("login name %s NOT authenticated, expected error", l.LoginName)
		} else if err != nil && err.Error() != "user not found" && err.Error() != "incorrect password" {
			t.Fatalf("user %s failed with error: %s", l.LoginName, err)
		}
	}
}

// GetJWTFromLoginName() is just running an 8-word SQL query to get a user ID
// and using go-chi jwtauth.New()
// not worth testing

// Upload profile pic
func TestHasAcceptableAspectRatio(t *testing.T) {
	var test_image_files = []struct {
		Name                     string
		HasAcceptableAspectRatio bool
	}{
		{"test1.webp", false},
		{"test2.webp", false},
		{"test3.webp", true},
	}

	_, user_test_file, _, _ := runtime.Caller(0)
	handler_util_dir := filepath.Dir(user_test_file)
	pic_dir := filepath.Join(handler_util_dir, "../../db/profile-pics")

	for _, l := range test_image_files {
		f, err := os.Open(pic_dir + "/" + l.Name)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		img, _, err := image.Decode(f)
		if err != nil {
			t.Fatal(err)
		}

		if HasAcceptableAspectRatio(img) != l.HasAcceptableAspectRatio {
			t.Fatalf("expected image %s to be %t", l.Name, l.HasAcceptableAspectRatio)
		}
	}
}
