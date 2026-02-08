package metrics

import (
	"database/sql"
	"log"
)

// Load Metrics
func LoadMetrics(db *sql.DB) {
	// get Pending jobs
	pending, err := CountPendingJobs(db)
	if err != nil {
		log.Fatal(err)
	}
	//get Success Jobs
	completedJobs, err := CountCompletedJobs(db)
	if err != nil {
		log.Fatal(err)
	}
	M.SuccessCount.Store(completedJobs)

	//get failed jobs
	failedJobs, err := CountFailedJobs(db)
	if err != nil {
		log.Fatal(err)
	}

	// get Total Retries
	totalRetry, err := CountTotalRetries(db)
	if err != nil {
		log.Fatal(err)
	}

	M.FailureCount.Store(failedJobs)
	M.DeadLetterCount.Store(failedJobs)
	M.QueueDepth.Store(pending)
	M.RetryCount.Store(totalRetry)
}

// queue Depth
func (m *Metrics) IncQueueDepth() {
	m.QueueDepth.Add(1)
}
func (m *Metrics) DecQueueDepth() {
	m.QueueDepth.Add(-1)
}

// Active Workers
func (m *Metrics) IncActiveWorkers() {
	m.ActiveWorkers.Add(1)
}
func (m *Metrics) DecActiveWorkers() {
	m.ActiveWorkers.Add(-1)
}

// Success Count
func (m *Metrics) IncSuccessCount() {
	m.SuccessCount.Add(1)
}

// Retry Count
func (m *Metrics) IncRetryCount() {
	m.RetryCount.Add(1)
}

// Failure Count
func (m *Metrics) IncFailureCount() {
	m.FailureCount.Add(1)
}

// DeadLetter Count
func (m *Metrics) IncDeadLetterCount() {
	m.DeadLetterCount.Add(1)
}

// Add Queue Time

func (m *Metrics) AddQueueTime(ms int64) {
	m.TotalQueueTimeMs.Add(ms)
}

// Add Execution Time

func (m *Metrics) AddExecutionTime(ms int64) {
	m.TotalExecutionTimeMs.Add(ms)
}
