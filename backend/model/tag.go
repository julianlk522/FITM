package model

import (
	"net/http"

	e "oitm/error"
	util "oitm/model/util"
)

type NewTag struct {
	LinkID string `json:"link_id"`
	Categories string `json:"categories"`
}

type NewTagRequest struct {
	*NewTag
	ID int64
	LastUpdated string
}

func (a *NewTagRequest) Bind(r *http.Request) error {
	if a.NewTag.Categories == "" {
		return e.ErrNoCats
	} else if a.NewTag.LinkID == "" {
		return e.ErrNoLinkID
	}

	a.LastUpdated = util.NEW_TIMESTAMP

	return nil
}

type EditTagRequest struct {
	ID string `json:"tag_id"`
	Categories string `json:"categories"`
	LastUpdated string
}

func (a *EditTagRequest) Bind(r *http.Request) error {
	if a.ID == "" {
		return e.ErrNoTagID
	}
	if a.Categories == "" {
		return e.ErrNoTagCategories
	}

	a.LastUpdated = util.NEW_TIMESTAMP

	return nil
}

// General
type Tag struct {
	ID string
	LinkID string
	Categories string
	SubmittedBy string
	LastUpdated string
}

type CategoryCount struct {
	Category string
	Count int32
}

func SortCategories(i, j CategoryCount) int {
	if i.Count > j.Count {
		return -1
	} else if i.Count == j.Count && i.Category < j.Category {
		return -1
	}
	return 1
}

// TagRanking is used to rank the top cats for a given link in order to
// calculate global cats. Tags are ranked by submit date.
type TagRanking struct {
	LifeSpanOverlap float32
	Categories string
}

type TagRankingPublic struct {
	TagRanking
	SubmittedBy string
	LastUpdated string
}

type TagPage struct {
	Link *LinkSignedIn
	UserTag *Tag
	TagRankings *[]TagRankingPublic
}