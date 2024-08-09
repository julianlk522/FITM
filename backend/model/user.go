package model

import (
	"net/http"

	e "oitm/error"
	util "oitm/model/util"
)

// AUTH
type Auth struct {
	LoginName string `json:"login_name"`
	Password string `json:"password"`
}
type SignUpRequest struct {
	*Auth
	CreatedAt string
}

func (a *SignUpRequest) Bind(r *http.Request) error {
	if a.Auth.LoginName == "" {
		return e.ErrNoLoginName
	} else if a.Auth.Password == "" {
		return e.ErrNoPassword
	}

	a.CreatedAt = util.NEW_TIMESTAMP
	return nil
}

type LogInRequest struct {
	*Auth
}


func (a *LogInRequest) Bind(r *http.Request) error {
	if a.Auth.LoginName == "" {
		return e.ErrNoLoginName
	} else if a.Auth.Password == "" {
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

type EditProfilePicRequest struct {
	ProfilePic string `json:"pfp,omitempty"`
}



// TREASURE MAP
type TreasureMapSections[T TmapLink | TmapLinkSignedIn] struct {
	Submitted *[]T
	Tagged *[]T
	Copied *[]T
	Categories *[]CategoryCount
}

type TreasureMap[T TmapLink | TmapLinkSignedIn] struct {
	Profile *Profile
	*TreasureMapSections[T]
}

type FilteredTreasureMap[T TmapLink | TmapLinkSignedIn] struct {
	*TreasureMapSections[T]
}