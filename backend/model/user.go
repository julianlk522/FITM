package model

import (
	"net/http"

	e "oitm/error"
	util "oitm/model/util"
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
	if a.UserAuth.LoginName == "" {
		return e.ErrNoLoginName
	} else if a.UserAuth.Password == "" {
		return e.ErrNoPassword
	}

	a.Created = util.NEW_TIMESTAMP
	return nil
}

type LogInRequest struct {
	*UserAuth
}


func (a *LogInRequest) Bind(r *http.Request) error {
	if a.UserAuth.LoginName == "" {
		return e.ErrNoLoginName
	} else if a.UserAuth.Password == "" {
		return e.ErrNoPassword
	}

	return nil
}



// PROFILE
type Profile struct {
	LoginName string
	About string
	PFP string
	Created string
}

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