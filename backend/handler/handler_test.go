package handler

import (
	"testing"

	"github.com/julianlk522/fitm/db"

	_ "github.com/mattn/go-sqlite3"

	"database/sql"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

const (
	test_user_id    = "3"
	test_login_name = "goolian"
)

var TestClient *sql.DB

func TestMain(m *testing.M) {
	var err error
	TestClient, err = sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		log.Fatal(err)
	}
	defer TestClient.Close()

	db.Client = TestClient
	log.Printf("switched connection to in-memory test DB")

	_, handler_test_file, _, _ := runtime.Caller(0)
	handler_dir := filepath.Dir(handler_test_file)
	db_dir := filepath.Join(handler_dir, "../db")
	sql_dump_path := filepath.Join(db_dir, "_fitm.db.sql")
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

	m.Run()
}
