package db

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"database/sql"
	"log"
)

func TestMain(m *testing.M) {
	var err error
	Client, err = sql.Open("sqlite3", "./oitm.db")
	if err != nil {
		log.Fatalf("Could not open database: %s", err)
	}

	m.Run()
}

func TestConnect(t *testing.T) {
	err := Client.Ping()
	if err != nil {
		t.Fatal(err)
	}
}