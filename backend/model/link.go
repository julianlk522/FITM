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
}

func (a *NewLinkRequest) Bind(r *http.Request) error {
	if a.NewLink == nil {
		return errors.New("missing required Link fields")
	}

	a.SubmitDate = time.Now().Format("2006-01-02 15:04:05")
	a.LikeCount = 0

	return nil
}

type LinkCopyRequest struct {
	ID int64
	LinkID string `json:"link_id"`
}

func (a *LinkCopyRequest) Bind(r *http.Request) error {
	if a.LinkID == "" {
		return errors.New("missing link ID")
	}

	return nil
}

type DeleteLinkCopyRequest struct {
	ID string `json:"copy_id"`
}

func (a *DeleteLinkCopyRequest) Bind(r *http.Request) error {
	if a.ID == "" {
		return errors.New("missing copy ID")
	}

	return nil
}

type Link struct {
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
	Link
	IsLiked bool
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