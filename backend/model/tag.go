package model

import (
	"net/http"
	"strconv"
	"strings"

	e "oitm/error"
	util "oitm/model/util"
)

type Tag struct {
	ID          string
	LinkID      string
	Categories  string
	SubmittedBy string
	LastUpdated string
}

type CatCount struct {
	Category string
	Count    int32
}

func SortCategories(i, j CatCount) int {
	if i.Count > j.Count {
		return -1
	} else if i.Count == j.Count && strings.ToLower(i.Category) < strings.ToLower(j.Category) {
		return -1
	}
	return 1
}

type TagRanking struct {
	LifeSpanOverlap float32
	Categories      string
}

type TagRankingPublic struct {
	TagRanking
	SubmittedBy string
	LastUpdated string
}

type TagPage struct {
	Link        *LinkSignedIn
	UserTag     *Tag
	TagRankings *[]TagRankingPublic
}

type NewTag struct {
	LinkID     string `json:"link_id"`
	Categories string `json:"categories"`
}

type NewTagRequest struct {
	*NewTag
	ID          string
	LastUpdated string
}

func (t *NewTagRequest) Bind(r *http.Request) error {
	if t.NewTag.LinkID == "" {
		return e.ErrNoLinkID
	} else if i, err := strconv.Atoi(t.NewTag.LinkID); err != nil || i < 1 {
		return e.ErrInvalidLinkID
	}

	switch {
		case t.NewTag.Categories == "":
			return e.ErrNoCats
		case util.HasTooLongCats(t.NewTag.Categories):
			return e.CatCharsExceedLimit(util.CAT_CHAR_LIMIT)
		case util.HasTooManyCats(t.NewTag.Categories):
			return e.NumCatsExceedsLimit(util.NUM_CATS_LIMIT)
		case util.HasDuplicateCats(t.NewTag.Categories):
			return e.ErrDuplicateCats
	}

	t.ID = util.NEW_UUID
	t.LastUpdated = util.NEW_LONG_TIMESTAMP

	return nil
}

type EditTagRequest struct {
	ID          string `json:"tag_id"`
	Categories  string `json:"categories"`
	LastUpdated string
}

func (et *EditTagRequest) Bind(r *http.Request) error {
	if et.ID == "" {
		return e.ErrNoTagID
	}

	switch {
		case et.Categories == "":
			return e.ErrNoCats
		case util.HasTooLongCats(et.Categories):
			return e.CatCharsExceedLimit(util.CAT_CHAR_LIMIT)
		case util.HasTooManyCats(et.Categories):
			return e.NumCatsExceedsLimit(util.NUM_CATS_LIMIT)
		case util.HasDuplicateCats(et.Categories):
			return e.ErrDuplicateCats
	}

	et.LastUpdated = util.NEW_LONG_TIMESTAMP

	return nil
}
