package model

import (
	"errors"
	"net/http"
	"time"
)

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
type EditProfileRequest struct {
	*EditAboutRequest
	*EditPfpRequest
}

func (a *EditProfileRequest) Bind(r *http.Request) error {
	if a.EditAboutRequest == nil && a.EditPfpRequest == nil {
		return errors.New("no data provided")
	}
	return nil
}

type EditAboutRequest struct {
	About string `json:"about,omitempty"`
}

type EditPfpRequest struct {
	PFP string `json:"pfp,omitempty"`
}

// TREASURE MAP

// end goal:
// type TreasureMap struct {
// 	Submitted []Link
// 	Tagged []Link
// 	Copied []Link
// 	Categories []CategoryCount
// }

type TreasureMap struct {
	Links []Link
	Categories []CategoryCount
}

// type TreasureMapSignedIn struct {
// 	Links []LinkSignedIn
// 	Categories []CategoryCount
// }

type TreasureMapSignedIn struct {
	Submitted []LinkSignedIn
	Tagged []LinkSignedIn
	Copied []LinkSignedIn
	Categories []CategoryCount
}

// GENERAL
type User struct {
	LoginName string
	About string
	PFP string
	Created string
}