package db

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"database/sql"
	"log"
)

func TestMain(m *testing.M) {
	var err error
	Client, err = sql.Open("sqlite-spellfix1", "./fitm.db")
	if err != nil {
		log.Fatalf("Could not open database with spellfix extension: %s", err)
	}

	m.Run()
}

func TestConnect(t *testing.T) {
	err := Client.Ping()
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoadSpellfix(t *testing.T) {
	_, err := Client.Exec(`SELECT word, rank FROM global_cats_spellfix;`)
	if err != nil {
		t.Fatal(err)
	}
}
