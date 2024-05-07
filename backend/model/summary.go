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
	if a.NewSummaryRequest == nil && a.NewSummaryLikeRequest == nil && a.EditSummaryRequest == nil && a.DeleteSummaryRequest == nil && a.DeleteSummaryLikeRequest == nil {
		return errors.New("missing required Summary fields")
	}

	if a.EditSummaryRequest != nil {
		if a.EditSummaryRequest.Text == "" {
			return errors.New("missing replacement summary text")
		} else if a.EditSummaryRequest.SummaryID == "" {
			return errors.New("missing summary ID")
		}
	}

	return nil
}

type NewSummaryRequest struct {
	SubmittedBy string `json:"submitted_by"`
	LinkID string `json:"link_id"`
	Text string `json:"text"`
}

type EditSummaryRequest struct {
	// would use json:"summary_id" here but conflicts with
	// below NewSummaryLikeRequest json ... not sure how else to fix
	SummaryID string `json:"summary_id_edit"`
	Text string `json:"text_edit"`
}

type DeleteSummaryRequest struct {
	// would use json:"summary_id" here but conflicts with
	// below NewSummaryLikeRequest json ... not sure how else to fix
	SummaryID string `json:"summary_id_del"`
}

type NewSummaryLikeRequest struct {
	SummaryID string `json:"summary_id"`
	UserID string `json:"user_id"`
}

type DeleteSummaryLikeRequest struct {
	SummaryLikeID string `json:"slike_id"`
}