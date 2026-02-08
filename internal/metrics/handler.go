package metrics

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// Metrics handler
func MetricsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// method check
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	finished := M.SuccessCount.Load() + M.FailureCount.Load()

	var avgQueueTime int64 = 0
	var avgExecutionTime int64 = 0

	if finished > 0 {
		avgQueueTime = M.TotalQueueTimeMs.Load() / finished
		avgExecutionTime = M.TotalExecutionTimeMs.Load() / finished
	}

	res := map[string]any{
		"queue_depth":          M.QueueDepth.Load(),
		"success_count":        M.SuccessCount.Load(),
		"retry_count":          M.RetryCount.Load(),
		"failure_count":        M.FailureCount.Load(),
		"dead_letter_count":    M.DeadLetterCount.Load(),
		"active_workers":       M.ActiveWorkers.Load(),
		"avg_time_in_queue_ms": avgQueueTime,
		"avg_execution_ms":     avgExecutionTime,
	}

	//responce
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)

}
