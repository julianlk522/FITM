package db

import (
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

		sql_dump_path := filepath.Join(test_data_path, "fitm_test.db.sql")
		sql_dump, err := os.ReadFile(sql_dump_path)
		if err != nil {
			t.Fatal(err)
		}
		_, err = Client.Exec(string(sql_dump))
		if err != nil {
			t.Fatal(err)
		}
		log.Print("loaded test data (TestLoadSpellfix)")

		// debug schema configuration
		rows, err := Client.Query(`
			SELECT name FROM sqlite_master 
			WHERE type='table';
		`)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		log.Println("Tables in database:")
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				t.Fatal(err)
			}
			log.Println(name)
		}
	}

	if _, err := Client.Exec(`SELECT word, rank FROM global_cats_spellfix LIMIT 1;`); err != nil {
		t.Fatal(err)
	}
}
