package main

import (
	"database/sql"
	"net/http"
)

func NewRouter(db *sql.DB) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/createJob", func(w http.ResponseWriter, r *http.Request) {
		CreatejobRequest(w, r, db)
	})

	return mux
}
