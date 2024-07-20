package handler

import (
	"database/sql"
	"errors"
	"fmt"

	"oitm/db"
)

const (
	LINKS_PAGE_LIMIT int = 20

	CATEGORY_PAGE_LIMIT int = 15
	CATEGORY_CONTRIBUTORS_LIMIT int = 5

	TAGS_PAGE_LIMIT int = 20
	NEW_TAG_CATEGORY_LIMIT int = 5
	TOP_OVERLAP_SCORES_LIMIT int = 20

	TMAP_CATEGORY_COUNT_LIMIT int = 5
)

var (
	DBClient *sql.DB = db.Client

	ErrNoPeriod error = errors.New("no period provided")
	ErrNoCategories error = errors.New("no categories provided")
	ErrNoLinkID error = errors.New("no link ID provided")
	ErrNoSummaryID error = errors.New("no summary ID provided")
	ErrNoLoginName error = errors.New("no login name provided")

	ErrNoLinkWithID error = errors.New("no link found with given ID")
	ErrNoSummaryWithID error = errors.New("no summary found with given ID")
	ErrNoTagWithID error = errors.New("no tag found with given ID")
	ErrNoUserWithLoginName error = errors.New("no user found with given login name")

	ErrTooManyCategories error = fmt.Errorf("too many tag categories (%d max)", NEW_TAG_CATEGORY_LIMIT)
	ErrDuplicateTag error = errors.New("duplicate tag")
	ErrDoesntOwnTag error = errors.New("cannot edit another user's tag")
)