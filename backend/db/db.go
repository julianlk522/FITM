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

func init() {
	if err := Connect(); err != nil {
		log.Fatal(err)
	}
}

func Connect() error {
	var err error

	_, db_file, _, _ := runtime.Caller(0)
	db_dir := filepath.Dir(db_file)

	Client, err = sql.Open("sqlite3", db_dir + "/oitm.db")
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