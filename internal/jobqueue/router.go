package jobqueue

import (
	"database/sql"
	"net/http"
)

func NewRouter(db *sql.DB, requestLimiter chan struct{}, producerLimiter chan struct{}) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/createJob", func(w http.ResponseWriter, r *http.Request) {
		CreatejobRequest(w, r, db, requestLimiter, producerLimiter)
	})

	return mux
}
