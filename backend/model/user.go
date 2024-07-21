package model

import (
	"errors"
	"net/http"
	"time"
)

type User struct {
	LoginName string
	About string
	PFP string
	Created string
}

// AUTH
type UserAuth struct {
	LoginName string `json:"login_name"`
	Password string `json:"password"`
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

// EDIT PROFILE
type EditAboutRequest struct {
	About string `json:"about"`
}

func (a *EditAboutRequest) Bind(r *http.Request) error {
	return nil
}

type EditPfpRequest struct {
	PFP string `json:"pfp,omitempty"`
}

// TREASURE MAP
type TreasureMap[T TmapLinkSignedIn | TmapLinkSignedOut] struct {
	Submitted *[]T
	Tagged *[]T
	Copied *[]T
	Categories *[]CategoryCount
}