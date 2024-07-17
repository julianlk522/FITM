package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var Client *sql.DB

func init() {
	if err := Connect(); err != nil {
		log.Fatal(err)
	}
}

func Connect() error {
	var err error

	Client, err = sql.Open("sqlite3", "./db/oitm.db")
	if err != nil {
		return err
	}

	err = Client.Ping()
	if err != nil {
		return err
	}

	log.Print("Connected to database")

	return nil
}