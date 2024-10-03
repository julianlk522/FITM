package query

import (
	"database/sql"
	"log"
	"testing"

	"github.com/julianlk522/fitm/db"
	"github.com/julianlk522/fitm/dbtest"
)

var TestClient *sql.DB

func TestMain(m *testing.M) {
	if err := dbtest.SetupTestDB(); err != nil {
		log.Fatal(err)
	}
	// TestClient unneeded but helps to reiterate in tests that the DB connection is temporary
	TestClient = db.Client
	m.Run()
}

func TestGetCatsWithEscapedChars(t *testing.T) {
	var test_cats = struct {
		Cats            []string
		ExpectedResults []string
	}{
		Cats:            []string{
			"slash/slash/slash", 
			"c. vi.per", 
			"hsien-ko",
		},
		ExpectedResults: []string{
			`slash"/"slash"/"slash`, 
			`c"." vi"."per`, 
			`hsien"-"ko`,
		},
	}

	got := GetCatsWithEscapedChars(test_cats.Cats)
	for i, res := range got {
		if res != test_cats.ExpectedResults[i] {
			t.Fatalf("got %s, want %s", got, test_cats.ExpectedResults)
		}
	}
}