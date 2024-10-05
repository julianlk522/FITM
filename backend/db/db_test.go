package db

import (
	"testing"
)

func TestConnect(t *testing.T) {
	err := Client.Ping()
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoadSpellfix(t *testing.T) {
	var word, rank string
	if err := Client.QueryRow(`SELECT word, rank FROM global_cats_spellfix LIMIT 1;`).Scan(&word, &rank); err != nil {
		t.Fatal(err)
	}
}
