package db

import (
	"database/sql"
	"log"
	"path/filepath"
	"runtime"

	"github.com/mattn/go-sqlite3"
)

var (
	Client *sql.DB
)

const AUTO_SUMMARY_USER_ID = "ca39e263-2ac7-4d70-abc5-b9b8f1bff332"

var _, db_file, _, _ = runtime.Caller(0)
var db_dir = filepath.Dir(db_file)

func init() {
	if err := Connect(); err != nil {
		log.Fatal(err)
	}
}

func Connect() error {
	LoadSpellfix()

	var err error
	Client, err = sql.Open("sqlite-spellfix1", db_dir+"/fitm.db")
	if err != nil {
		return err
	}

	err = Client.Ping()
	if err != nil {
		return err
	}

	log.Print("DB connection verified")

	return nil
}

func LoadSpellfix() {
	sql.Register(
		"sqlite-spellfix1",
		&sqlite3.SQLiteDriver{
			ConnectHook: func(c *sqlite3.SQLiteConn) error {
				return c.LoadExtension(filepath.Join(db_dir, "spellfix"), "sqlite3_spellfix_init")
			},
		},
	)
	log.Print("Loaded spellfix")
}
