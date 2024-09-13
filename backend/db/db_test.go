package db

import (
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"database/sql"
	"log"
)

var (
	TestClient *sql.DB
)

func TestMain(m *testing.M) {
	var err error
	TestClient, err = sql.Open("sqlite-spellfix1", "./fitm.db")
	if err != nil {
		log.Fatalf("Could not open database file with spellfix extension: %s", err)
	}

	// switch to in-memory DB connection
	TestClient, err = sql.Open("sqlite-spellfix1", "file::memory:?cache=shared")
	if err != nil {
		log.Fatalf("Could switch to in-memory database: %s", err)
	}
	log.Printf("switched connection to in-memory test DB (db_test.go)")

	var sql_dump_path string

	// check for FITM_TEST_DATA_PATH env var,
	// if not set, use default path
	test_data_path := os.Getenv("FITM_TEST_DATA_PATH")
	if test_data_path == "" {
		log.Printf("FITM_TEST_DATA_PATH not set, using default path")
		sql_dump_path = filepath.Join(db_dir, "fitm_test.db.sql")
	} else {
		log.Print("using FITM_TEST_DATA_PATH")
		sql_dump_path = test_data_path + "/fitm_test.db.sql"
	}

	sql_dump, err := os.ReadFile(sql_dump_path)
	if err != nil {
		log.Fatal(err)
	}
	_, err = TestClient.Exec(string(sql_dump))
	if err != nil {
		log.Fatal(err)
	}

	// verify that in-memory DB has new test data
	var link_id string
	err = TestClient.QueryRow("SELECT id FROM Links WHERE id = '1';").Scan(&link_id)
	if err != nil {
		log.Fatalf("in-memory DB did not receive dump data: %s", err)
	}
	log.Printf("verified dump data added to test DB")

	// verify spellfix working
	_, err = TestClient.Exec("SELECT word, rank FROM global_cats_spellfix;")
	if err != nil {
		log.Fatalf("in-memory DB did not receive spellfix: %s", err)
	}
	log.Printf("verified spellfix loaded into test DB")

	m.Run()
}

func TestConnect(t *testing.T) {
	err := TestClient.Ping()
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoadSpellfix(t *testing.T) {
	_, err := TestClient.Exec(`SELECT word, rank FROM global_cats_spellfix;`)
	if err != nil {
		t.Fatal(err)
	}
}
