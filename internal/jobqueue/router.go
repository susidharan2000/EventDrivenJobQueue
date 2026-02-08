package jobqueue

import (
	"database/sql"
	"net/http"

	"github.com/susi/EventDrivenJobQueue/internal/metrics"
)

func NewRouter(db *sql.DB, requestLimiter chan struct{}, producerLimiter chan struct{}) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/createJob", func(w http.ResponseWriter, r *http.Request) {
		CreatejobRequest(w, r, db, requestLimiter, producerLimiter)
	})

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics.MetricsHandler(w, r, db)
	})

	return mux
}
