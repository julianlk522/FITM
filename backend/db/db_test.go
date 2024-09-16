package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestConnect(t *testing.T) {
	err := Client.Ping()
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoadSpellfix(t *testing.T) {

	// load test data if FITM_TEST_DATA_PATH is set
	// (cannot run dbtest.SetupTestDB here since importing creates a circular dependency)
	test_data_path := os.Getenv("FITM_TEST_DATA_PATH")
	if test_data_path != "" {
		log.Print("FITM_TEST_DATA_PATH found: loading test data to run spellfix check against")
		
		Client, err := sql.Open("sqlite-spellfix1", "file::memory:?cache=shared")
		if err != nil {
			t.Fatalf("could not open in-memory DB: %s", err)
		}
		sql_dump_path := filepath.Join(db_dir, "fitm_test.db.sql")
		sql_dump, err := os.ReadFile(sql_dump_path)
		if err != nil {
			t.Fatal(err)
		}
		_, err = Client.Exec(string(sql_dump))
		if err != nil {
			t.Fatal(err)
		}
		log.Print("loaded test data (TestLoadSpellfix)")
	}
	_, err := Client.Exec(`SELECT word, rank FROM global_cats_spellfix;`)
	if err != nil {
		t.Fatal(err)
	}
}
