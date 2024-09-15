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
	_, err := Client.Exec(`SELECT word, rank FROM global_cats_spellfix;`)
	if err != nil {
		t.Fatal(err)
	}
}
