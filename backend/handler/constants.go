package handler

import (
	"database/sql"
	"errors"

	"oitm/db"
)

var (
	DBClient *sql.DB = db.Client

	ErrNoPeriod error = errors.New("no period provided")
	ErrNoCategories error = errors.New("no categories provided")
	ErrNoLinkID error = errors.New("no link ID provided")
	ErrInvalidLinkID error = errors.New("invalid link ID provided")
	ErrNoLinkWithID error = errors.New("no link found with given ID")
)

const (
	LINKS_PAGE_LIMIT int = 20
	CATEGORY_CONTRIBUTORS_LIMIT int = 5
	CATEGORY_COUNT_LIMIT int = 20
	NEW_TAG_CATEGORY_LIMIT int = 5
	TOP_TAG_CATEGORIES_LIMIT int = 15
)