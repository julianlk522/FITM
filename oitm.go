package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("REST API for Open Internet Treasure Map")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Open Internet Treasure Map")
	})
	mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World!")
	})

	if err := http.ListenAndServe("localhost:8000", mux); err != nil {
		fmt.Println(err.Error())
	}
}
