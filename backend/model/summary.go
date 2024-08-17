package model

import (
	"net/http"
	"strconv"

	e "oitm/error"
	util "oitm/model/util"

	"github.com/google/uuid"
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
	ID string
	LinkID string `json:"link_id"`
	Text string `json:"text"`
	LastUpdated string
}

func (s *NewSummaryRequest) Bind(r *http.Request) error {
	if s.LinkID == "" {
		return e.ErrNoLinkID
	} else if i, err := strconv.Atoi(s.LinkID); err != nil || i < 1 {
		return e.ErrInvalidLinkID
	}
	
	if s.Text == "" {
		return e.ErrNoSummaryText
	} else if len(s.Text) > util.SUMMARY_CHAR_LIMIT {
		return e.SummaryLengthExceedsLimit(util.SUMMARY_CHAR_LIMIT)
	}

	s.ID = uuid.New().String()
	s.LastUpdated = util.NEW_TIMESTAMP
	
	return nil

}

// DELETE
type DeleteSummaryRequest struct {
	SummaryID string `json:"summary_id"`
}

func (ds *DeleteSummaryRequest) Bind(r *http.Request) error {
	if ds.SummaryID == "" {
		return e.ErrNoSummaryID
	}
	return nil
}

// EDIT
type EditSummaryRequest struct {
	SummaryID string `json:"summary_id"`
	Text string `json:"text"`
}

func (es *EditSummaryRequest) Bind(r *http.Request) error {
	if es.SummaryID == "" {
		return e.ErrNoSummaryID
	}
	if es.Text == "" {
		return e.ErrNoSummaryReplacementText
	}
	return nil
}