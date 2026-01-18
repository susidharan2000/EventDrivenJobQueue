package jobqueue

import (
	"encoding/json"
	"time"
)

type CreateJob struct {
	Type           string          `json:"type"`
	Status         string          `json:"status"`
	Payload        json.RawMessage `json:"payload"`
	MaxRetries     int             `json:"max_retries"`
	RunAt          time.Time       `json:"run_at"`
	IdempotencyKey *string         `json:"idempotency_key"`
}

type WorkerJob struct {
	Id         int             `json:"id"`
	Type       string          `json:"type"`
	Status     string          `json:"status"`
	Payload    json.RawMessage `json:"payload"`
	MaxRetries int             `json:"max_retries"`
	Attempts   int             `json:"attempts"`
	RunAt      time.Time       `json:"run_at"`
}

type Email struct {
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}
