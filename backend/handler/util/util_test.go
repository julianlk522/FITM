package handler

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/julianlk522/fitm/db"

	_ "github.com/mattn/go-sqlite3"
)

// shared across handler/util tests
var (
	TestClient *sql.DB

	test_login_name = "jlk"
	test_user_id    = "3"

	test_req_login_name = "bradley"
	test_req_user_id    = "13"

	test_single_cat    = []string{"umvc3"}
	test_multiple_cats = []string{"umvc3", "flowers"}

	test_link_id = "1"
)

func TestMain(m *testing.M) {
	var err error
	TestClient, err = sql.Open("sqlite-spellfix1", "file::memory:?cache=shared")
	if err != nil {
		log.Fatal(err)
	}
	defer TestClient.Close()

	db.Client = TestClient
	log.Printf("switched connection to in-memory test DB (util_test.go)")

	var sql_dump_path string
	// check for FITM_TEST_DATA_PATH env var,
	// if not set, use default path
	test_data_path := os.Getenv("FITM_TEST_DATA_PATH")
	if test_data_path == "" {
		log.Printf("FITM_TEST_DATA_PATH not set, using default path")
		
		_, util_file, _, _ := runtime.Caller(0)
		util_dir := filepath.Dir(util_file)
		db_dir := filepath.Join(util_dir, "../../db")
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
	err = TestClient.QueryRow("SELECT id FROM Links WHERE id = ?;", test_link_id).Scan(&link_id)
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
