package query

import (
	"database/sql"

	"github.com/julianlk522/fitm/db"
)

// shared across Query tests
var (
	TestClient *sql.DB = db.Client
)