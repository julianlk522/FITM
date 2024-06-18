package model

import (
	"errors"
	"net/http"
	"time"
)

type NewLink struct {
	URL string `json:"url"`
	Categories string `json:"categories"`
}

type NewLinkRequest struct {
	*NewLink
	ID int64
	SubmittedBy string
	SubmitDate string
	Summary string
	SummaryCount int
	LikeCount int64
	ImgURL string

	// to be assigned by handler after processing
	URL string
	Categories string
}

func (a *NewLinkRequest) Bind(r *http.Request) error {
	if a.NewLink.URL == "" {
		return errors.New("missing url")
	} else if a.NewLink.Categories == "" {
		return errors.New("missing categories")
	}

	a.SubmitDate = time.Now().Format("2006-01-02 15:04:05")
	a.LikeCount = 0

	return nil
}

type LinkSignedOut struct {
	ID int64
	URL string
	SubmittedBy string
	SubmitDate string
	Categories string
	Summary string
	SummaryCount int
	LikeCount int64
	ImgURL string
}

type LinkSignedIn struct {
	LinkSignedOut
	IsLiked bool
	IsCopied bool
}

type Link interface {
	GetCategories() string
}

func (l LinkSignedOut) GetCategories() string {
	return l.Categories
}

func (l LinkSignedIn) GetCategories() string {
	return l.Categories
}

type CustomLinkCategories struct {
	LinkID int64
	Categories string
}

type CategoryContributor struct {
	Categories string
	LoginName string
	LinksSubmitted int
}