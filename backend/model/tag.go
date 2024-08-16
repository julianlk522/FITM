package model

import (
	"net/http"

	e "oitm/error"
	util "oitm/model/util"
)

type Tag struct {
	ID string
	LinkID string
	Categories string
	SubmittedBy string
	LastUpdated string
}

type CatCount struct {
	Category string
	Count int32
}

func SortCategories(i, j CatCount) int {
	if i.Count > j.Count {
		return -1
	} else if i.Count == j.Count && i.Category < j.Category {
		return -1
	}
	return 1
}

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



type NewTag struct {
	LinkID string `json:"link_id"`
	Categories string `json:"categories"`
}

type NewTagRequest struct {
	*NewTag
	ID int64
	LastUpdated string
}

func (t *NewTagRequest) Bind(r *http.Request) error {
	if t.NewTag.Categories == "" {
		return e.ErrNoCats
	} else if util.HasTooLongCats(t.NewTag.Categories) {
		return e.CatCharsExceedLimit(util.CAT_CHAR_LIMIT)
	} else if util.IsTooManyCats(t.NewTag.Categories) {
		return e.NumCatsExceedsLimit(util.NUM_CATS_LIMIT)
	}
	
	if t.NewTag.LinkID == "" {
		return e.ErrNoLinkID
	}

	t.LastUpdated = util.NEW_TIMESTAMP

	return nil
}

type EditTagRequest struct {
	ID string `json:"tag_id"`
	Categories string `json:"categories"`
	LastUpdated string
}

func (et *EditTagRequest) Bind(r *http.Request) error {
	if et.ID == "" {
		return e.ErrNoTagID
	}
	if et.Categories == "" {
		return e.ErrNoTagCats
	}

	et.LastUpdated = util.NEW_TIMESTAMP

	return nil
}