package model

import (
	"net/http"

	e "oitm/error"
	util "oitm/model/util"
)

type Summary struct {
	ID string
	Text string
	SubmittedBy string
	LastUpdated string
	LikeCount int
}

type SummarySignedIn struct {
	Summary
	IsLiked bool
}

type SummaryPage[S SummarySignedIn | Summary, L LinkSignedIn | Link ] struct {
	Link L
	Summaries []S
}

// ADD
type NewSummaryRequest struct {
	LinkID string `json:"link_id"`
	Text string `json:"text"`
	LastUpdated string
}

func (a *NewSummaryRequest) Bind(r *http.Request) error {
	if a.LinkID == "" {
		return e.ErrNoLinkID
	}
	if a.Text == "" {
		return e.ErrNoSummaryText
	}

	a.LastUpdated = util.NEW_TIMESTAMP
	
	return nil

}

// DELETE
type DeleteSummaryRequest struct {
	SummaryID string `json:"summary_id"`
}

func (a *DeleteSummaryRequest) Bind(r *http.Request) error {
	if a.SummaryID == "" {
		return e.ErrNoSummaryID
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
		return e.ErrNoSummaryID
	}
	if a.Text == "" {
		return e.ErrNoSummaryReplacementText
	}
	return nil
}