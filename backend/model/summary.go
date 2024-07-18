package model

import (
	"errors"
	"net/http"
	"time"
)

// ADD
type NewSummaryRequest struct {
	LinkID string `json:"link_id"`
	Text string `json:"text"`
	LastUpdated string
}

func (a *NewSummaryRequest) Bind(r *http.Request) error {
	if a.LinkID == "" {
		return errors.New("missing link ID")
	}
	if a.Text == "" {
		return errors.New("missing summary text")
	}

	a.LastUpdated = time.Now().Format("2006-01-02 15:04:05")
	
	return nil

}

// DELETE
type DeleteSummaryRequest struct {
	SummaryID string `json:"summary_id"`
}

func (a *DeleteSummaryRequest) Bind(r *http.Request) error {
	if a.SummaryID == "" {
		return errors.New("missing summary ID")
	}
	return nil
}

// EDIT
type EditSummaryRequest struct {
	SummaryID string `json:"summary_id"`
	Text string `json:"text"`
}

func (a *EditSummaryRequest) Bind(r *http.Request) error {
	if a.SummaryID == "" {
		return errors.New("missing summary ID")
	}
	if a.Text == "" {
		return errors.New("missing replacement text")
	}
	return nil
}

// GENERAL
type SummarySignedOut struct {
	ID string
	Text string
	SubmittedBy string
	LastUpdated string
	LikeCount int
}

type SummarySignedIn struct {
	SummarySignedOut
	IsLiked bool
}

type SummaryPage[S SummarySignedIn | SummarySignedOut, L LinkSignedIn | LinkSignedOut ] struct {
	Link L
	Summaries []S
}