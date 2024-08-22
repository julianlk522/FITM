package db

import (
	"database/sql"
	"log"
	"path/filepath"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
)

var (
	Client *sql.DB
)

const AUTO_SUMMARY_USER_ID = "ca39e263-2ac7-4d70-abc5-b9b8f1bff332"

func init() {
	if err := Connect(); err != nil {
		log.Fatal(err)
	}
}

func Connect() error {
	var err error

	_, db_file, _, _ := runtime.Caller(0)
	db_dir := filepath.Dir(db_file)

	Client, err = sql.Open("sqlite3", db_dir+"/oitm.db")
	if err != nil {
		return err
	}

	err = Client.Ping()
	if err != nil {
		return err
	}

	log.Print("Connected to DB")

	return nil
}
