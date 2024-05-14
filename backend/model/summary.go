package model

import (
	"errors"
	"net/http"
)

type SummaryRequest struct {
	*NewSummaryRequest
	*EditSummaryRequest
	*DeleteSummaryRequest
	*NewSummaryLikeRequest
	*DeleteSummaryLikeRequest
}

func (a *SummaryRequest) Bind(r *http.Request) error {
	// New Summary
	if a.NewSummaryRequest != nil {
		if a.NewSummaryRequest.SubmittedByID == "" {
			return errors.New("missing submitted_by_id")
		}
		if a.NewSummaryRequest.LinkID == "" {
			return errors.New("missing link_id")
		}
		if a.NewSummaryRequest.Text == "" {
			return errors.New("missing summary text")
		}

		// error if other fields present
		if a.EditSummaryRequest != nil || a.DeleteSummaryRequest != nil || a.NewSummaryLikeRequest != nil || a.DeleteSummaryLikeRequest != nil {
			return errors.New("invalid fields")
		}
	}

	// Edit Summary
	if a.EditSummaryRequest != nil {
		if a.EditSummaryRequest.SummaryID == "" {
			return errors.New("missing summary_id")
		}
		if a.EditSummaryRequest.Text == "" {
			return errors.New("missing replacement text")
		}

		// error if other fields present
		if a.NewSummaryRequest != nil || a.DeleteSummaryRequest != nil || a.NewSummaryLikeRequest != nil || a.DeleteSummaryLikeRequest != nil {
			return errors.New("invalid fields")
		}
	}

	// Delete Summary
	if a.DeleteSummaryRequest != nil {
		if a.DeleteSummaryRequest.SummaryID == "" {
			return errors.New("missing summary_id")
		}

		// error if other fields present
		if a.NewSummaryRequest != nil || a.EditSummaryRequest != nil || a.NewSummaryLikeRequest != nil || a.DeleteSummaryLikeRequest != nil {
			return errors.New("invalid fields")
		}
	}

	// New Summary Like
	if a.NewSummaryLikeRequest != nil {
		if a.NewSummaryLikeRequest.SummaryID == "" {
			return errors.New("missing summary_id")
		}
		if a.NewSummaryLikeRequest.UserID == "" {
			return errors.New("missing user_id")
		}

		// error if other fields present
		if a.NewSummaryRequest != nil || a.EditSummaryRequest != nil || a.DeleteSummaryRequest != nil || a.DeleteSummaryLikeRequest != nil {
			return errors.New("invalid fields")
		}
	}

	// Delete Summary Like
	if a.DeleteSummaryLikeRequest != nil {
		if a.DeleteSummaryLikeRequest.SummaryLikeID == "" {
			return errors.New("missing slike_id")
		}

		// error if other fields present
		if a.NewSummaryRequest != nil || a.EditSummaryRequest != nil || a.DeleteSummaryRequest != nil || a.NewSummaryLikeRequest != nil {
			return errors.New("invalid fields")
		}
	}

	return nil
}

type NewSummaryRequest struct {
	SubmittedByID string `json:"submitted_by_id"`
	LinkID string `json:"link_id"`
	Text string `json:"text"`
}

type EditSummaryRequest struct {
	SummaryID string `json:"summary_id_edit"`
	Text string `json:"text_edit"`
}

type DeleteSummaryRequest struct {
	SummaryID string `json:"summary_id_del"`
}

type NewSummaryLikeRequest struct {
	SummaryID string `json:"summary_id"`
	UserID string `json:"user_id"`
}

type DeleteSummaryLikeRequest struct {
	SummaryLikeID string `json:"slike_id"`
}

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