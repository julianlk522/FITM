package model

import (
	"net/http"
	e "oitm/error"

	util "oitm/model/util"
)

type Link struct {
	ID int64
	URL string
	SubmittedBy string
	SubmitDate string
	Categories string
	Summary string
	SummaryCount int
	TagCount int
	LikeCount int64
	ImgURL string
}

type LinkSignedIn struct {
	Link
	IsLiked bool
	IsTagged bool
	IsCopied bool
}

type PaginatedLinks[T Link | LinkSignedIn] struct {
	Links *[]T
	NextPage int
}



type TmapLink struct {
	Link
	CategoriesFromUser bool
}

type TmapLinkSignedIn struct {
	LinkSignedIn
	CategoriesFromUser bool
}

type CategoryContributor struct {
	Categories string
	LoginName string
	LinksSubmitted int
}

type NewLink struct {
	URL string `json:"url"`
	Categories string `json:"categories"`
	Summary string `json:"summary,omitempty"`
}

type NewLinkRequest struct {
	*NewLink
	SubmitDate string
	LikeCount int64
	
	// to be assigned by handler
	ID int64
	URL string // potentially modified after test request(s)
	SubmittedBy string
	Categories string // used after sort
	AutoSummary string
	SummaryCount int
	ImgURL string
}

func (a *NewLinkRequest) Bind(r *http.Request) error {
	if a.NewLink.URL == "" {
		return e.ErrNoURL
	}
	if a.NewLink.Categories == "" {
		return e.ErrNoTagCats
	} else if util.IsTooManyCats(a.NewLink.Categories) {
		return e.ErrTooManyCats
	}

	a.SubmitDate = util.NEW_TIMESTAMP
	a.LikeCount = 0

	return nil
}