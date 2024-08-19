package model

import (
	"net/http"
	e "oitm/error"

	util "oitm/model/util"
)

type Link struct {
	ID           string
	URL          string
	SubmittedBy  string
	SubmitDate   string
	Categories   string
	Summary      string
	SummaryCount int
	TagCount     int
	LikeCount    int64
	ImgURL       string
}

type LinkSignedIn struct {
	Link
	IsLiked  bool
	IsTagged bool
	IsCopied bool
}

type PaginatedLinks[T Link | LinkSignedIn] struct {
	Links    *[]T
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
	Categories     string
	LoginName      string
	LinksSubmitted int
}

type NewLink struct {
	URL        string `json:"url"`
	Categories string `json:"categories"`
	Summary    string `json:"summary,omitempty"`
}

type NewLinkRequest struct {
	*NewLink
	ID         string
	SubmitDate string
	LikeCount  int64

	// to be assigned by handler
	URL          string // potentially modified after test request(s)
	SubmittedBy  string
	Categories   string // potentially modified after sort
	AutoSummary  string
	SummaryCount int
	ImgURL       string
}

func (l *NewLinkRequest) Bind(r *http.Request) error {
	if l.NewLink.URL == "" {
		return e.ErrNoURL
	} else if len(l.NewLink.URL) > util.URL_CHAR_LIMIT {
		return e.LinkURLCharsExceedLimit(util.URL_CHAR_LIMIT)
	}

	switch {
		case l.NewLink.Categories == "":
			return e.ErrNoTagCats
		case util.HasTooLongCats(l.NewLink.Categories):
			return e.CatCharsExceedLimit(util.CAT_CHAR_LIMIT)
		case util.HasTooManyCats(l.NewLink.Categories):
			return e.NumCatsExceedsLimit(util.NUM_CATS_LIMIT)
		case util.HasDuplicateCats(l.NewLink.Categories):
			return e.ErrDuplicateCats
	}

	if len(l.NewLink.Summary) > util.SUMMARY_CHAR_LIMIT {
		return e.SummaryLengthExceedsLimit(util.SUMMARY_CHAR_LIMIT)
	}

	l.ID = util.NEW_UUID
	l.SubmitDate = util.NEW_LONG_TIMESTAMP
	l.LikeCount = 0

	return nil
}
