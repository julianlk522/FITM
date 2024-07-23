package handler

import (
	"database/sql"

	"oitm/db"
)

const (
	LINKS_PAGE_LIMIT int = 20
	CATEGORY_PAGE_LIMIT int = 15
	TAGS_PAGE_LIMIT int = 20

	TOP_OVERLAP_SCORES_LIMIT int = 20
	
	CATEGORY_CONTRIBUTORS_LIMIT int = 5
	TMAP_CATEGORY_COUNT_LIMIT int = 5
)

var (
	DBClient *sql.DB = db.Client
)