package model

import (
	"errors"
	"net/http"
)

// ADD
type NewSummaryOrSummaryLikeRequest struct {
	*NewSummaryRequest
	*NewSummaryLikeRequest
}

func (a *NewSummaryOrSummaryLikeRequest) Bind(r *http.Request) error {
	if a.NewSummaryRequest != nil {
		return a.NewSummaryRequest.Bind(r)
	}
	return a.NewSummaryLikeRequest.Bind(r)
}

type NewSummaryRequest struct {
	LinkID string `json:"link_id"`
	Text string `json:"text"`
}

func (a *NewSummaryRequest) Bind(r *http.Request) error {
	if a.LinkID == "" {
		return errors.New("missing link ID")
	}
	if a.Text == "" {
		return errors.New("missing summary text")
	}
	return nil
}

type NewSummaryLikeRequest struct {
	SummaryID string `json:"summary_id"`
}

func (a *NewSummaryLikeRequest) Bind(r *http.Request) error {
	if a.SummaryID == "" {
		return errors.New("missing summary ID")
	}
	return nil
}

// DELETE
type DeleteSummaryOrSummaryLikeRequest struct {
	*DeleteSummaryRequest
	*DeleteSummaryLikeRequest
}

func (a *DeleteSummaryOrSummaryLikeRequest) Bind(r *http.Request) error {
	if a.DeleteSummaryRequest != nil {
		return a.DeleteSummaryRequest.Bind(r)
	}
	return a.DeleteSummaryLikeRequest.Bind(r)
}

type DeleteSummaryRequest struct {
	SummaryID string `json:"summary_id"`
}

func (a *DeleteSummaryRequest) Bind(r *http.Request) error {
	if a.SummaryID == "" {
		return errors.New("missing summary ID")
	}
	return nil
}

type DeleteSummaryLikeRequest struct {
	SummaryLikeID string `json:"slike_id"`
}

func (a *DeleteSummaryLikeRequest) Bind(r *http.Request) error {
	if a.SummaryLikeID == "" {
		return errors.New("missing summary like ID")
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
	LikeCount int
}

type SummaryPage struct {
	Link Link
	Summaries []Summary
}