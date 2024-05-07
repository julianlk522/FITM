package model

import (
	"errors"
	"net/http"
	"time"
)

type UserAuth struct {
	LoginName string `json:"login_name"`
	Password string `json:"password"`
}

type User struct {
	*UserAuth
	ID int64
	About string
	ProfilePic string
}
type SignUpRequest struct {
	*UserAuth
	Created string
}

func (a *SignUpRequest) Bind(r *http.Request) error {
	if a.UserAuth == nil {
		return errors.New("signup info not provided")
	} else if a.UserAuth.LoginName == "" {
		return errors.New("missing login name")
	} else if a.UserAuth.Password == "" {
		return errors.New("missing password")
	}

	a.Created = time.Now().Format("2006-01-02 15:04:05")
	return nil
}

type LogInRequest struct {
	*UserAuth
}


func (a *LogInRequest) Bind(r *http.Request) error {
	if a.UserAuth == nil {
		return errors.New("login info not provided")
	} else if a.UserAuth.LoginName == "" {
		return errors.New("missing login name")
	} else if a.UserAuth.Password == "" {
		return errors.New("missing password")
	}

	return nil
}

type EditProfileRequest struct {
	AuthToken string `json:"token"`
	*EditAboutRequest
	*EditPfpRequest
}

func (a *EditProfileRequest) Bind(r *http.Request) error {
	if a.AuthToken == "" {
		return errors.New("missing auth token")
	}

	// TODO: check auth token

	return nil
}

type EditAboutRequest struct {
	About string `json:"about,omitempty"`
}

type EditPfpRequest struct {
	PFP string `json:"pfp,omitempty"`
}