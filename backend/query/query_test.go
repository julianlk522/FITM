package query

import (
	"database/sql"
	"oitm/db"
)

// shared across Query tests
var (
	TestClient *sql.DB = db.Client
)