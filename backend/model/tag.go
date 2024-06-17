package model

import (
	"errors"
	"net/http"
	"time"
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
	if a.NewTag == nil {
		return errors.New("missing required Tag fields")
	}

	a.LastUpdated = time.Now().Format("2006-01-02 15:04:05")

	return nil
}

type EditTagRequest struct {
	ID string `json:"tag_id"`
	Categories string `json:"categories"`
	LastUpdated string
}

func (a *EditTagRequest) Bind(r *http.Request) error {
	if a.ID == "" {
		return errors.New("missing tag ID")
	}
	if a.Categories == "" {
		return errors.New("missing categories")
	}

	a.LastUpdated = time.Now().Format("2006-01-02 15:04:05")

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

type EarlyTag struct {
	LifeSpanOverlap float32
	Categories string
}

type EarlyTagPublic struct {
	EarlyTag
	SubmittedBy string
	LastUpdated string
}

type TagPage struct {
	Link *LinkSignedIn
	UserTag *Tag
	TopTags *[]EarlyTagPublic
}