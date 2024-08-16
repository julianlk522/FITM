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

func (s *SignUpRequest) Bind(r *http.Request) error {
	if s.Auth.LoginName == "" {
		return e.ErrNoLoginName
	} else if len(s.Auth.LoginName) < util.LOGIN_NAME_LOWER_LIMIT {
		return e.LoginNameExceedsLowerLimit(util.LOGIN_NAME_LOWER_LIMIT)
	} else if len(s.Auth.LoginName) > util.LOGIN_NAME_UPPER_LIMIT {
		return e.LoginNameExceedsUpperLimit(util.LOGIN_NAME_UPPER_LIMIT)
	} else if s.Auth.Password == "" {
		return e.ErrNoPassword
	} else if len(s.Auth.Password) < util.PASSWORD_LOWER_LIMIT {
		return e.PasswordExceedsLowerLimit(util.PASSWORD_LOWER_LIMIT)
	} else if len(s.Auth.Password) > util.PASSWORD_UPPER_LIMIT {
		return e.PasswordExceedsUpperLimit(util.PASSWORD_UPPER_LIMIT)
	}

	s.CreatedAt = util.NEW_TIMESTAMP
	return nil
}

type LogInRequest struct {
	*Auth
}


func (l *LogInRequest) Bind(r *http.Request) error {
	if l.Auth.LoginName == "" {
		return e.ErrNoLoginName
	} else if l.Auth.Password == "" {
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

func (ea *EditAboutRequest) Bind(r *http.Request) error {
	if len(ea.About) > util.PROFILE_ABOUT_CHAR_LIMIT {
		return e.ProfileAboutLengthExceedsLimit(util.PROFILE_ABOUT_CHAR_LIMIT)
	}

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
	Categories *[]CatCount
}

type TreasureMap[T TmapLink | TmapLinkSignedIn] struct {
	Profile *Profile
	*TreasureMapSections[T]
}

type FilteredTreasureMap[T TmapLink | TmapLinkSignedIn] struct {
	*TreasureMapSections[T]
}