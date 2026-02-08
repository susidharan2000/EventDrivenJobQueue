package metrics

import "sync/atomic"

type Metrics struct {
	QueueDepth      atomic.Int64
	SuccessCount    atomic.Int64
	RetryCount      atomic.Int64
	FailureCount    atomic.Int64
	DeadLetterCount atomic.Int64
	ActiveWorkers   atomic.Int64

	TotalQueueTimeMs     atomic.Int64
	TotalExecutionTimeMs atomic.Int64
}

var M *Metrics = &Metrics{}
