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
	SubmitDate string
}

func (a *NewSummaryRequest) Bind(r *http.Request) error {
	if a.LinkID == "" {
		return errors.New("missing link ID")
	}
	if a.Text == "" {
		return errors.New("missing summary text")
	}

	a.SubmitDate = time.Now().Format("2006-01-02 15:04:05")
	
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
type Summary struct {
	ID string
	Text string
	SubmittedBy string
	SubmitDate string
	LikeCount int
}

type SummarySignedIn struct {
	ID string
	Text string
	SubmittedBy string
	SubmitDate string
	LikeCount int
	IsLiked bool
}

type SummaryPage struct {
	Link Link
	Summaries []Summary
}

type SummaryPageSignedIn struct {
	Link LinkSignedIn
	Summaries []SummarySignedIn
}