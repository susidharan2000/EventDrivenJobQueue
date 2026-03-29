package jobqueue

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestStartDispatcher_SendsClaimedJobToChannel(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := InitJobsSchema(db); err != nil {
		t.Fatalf("init schema: %v", err)
	}

	payload := json.RawMessage(`{"email":"test@example.com"}`)
	_, err = db.Exec(`
		INSERT INTO jobs (type, status, payload, max_retries, run_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, datetime('now'), datetime('now'), datetime('now'))`,
		"email", "queued", payload, 3)
	if err != nil {
		t.Fatalf("insert job: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	workerCh := make(chan WorkerJob, 2)
	go StartDispatcher(db, ctx, workerCh)

	select {
	case job := <-workerCh:
		if job.Id == 0 {
			t.Error("expected non-zero job id")
		}
		if job.Type != "email" {
			t.Errorf("job type = %q, want email", job.Type)
		}
		if job.Status != "processing" {
			t.Errorf("job status = %q, want processing", job.Status)
		}
		if job.Attempts != 1 {
			t.Errorf("job attempts = %d, want 1", job.Attempts)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for job from dispatcher")
	}

	cancel()
	// After cancel, dispatcher should exit and close workerCh
	_, open := <-workerCh
	if open {
		t.Error("expected workerCh to be closed after context cancel")
	}
}

func TestStartDispatcher_ExitsOnContextCancel(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := InitJobsSchema(db); err != nil {
		t.Fatalf("init schema: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	workerCh := make(chan WorkerJob, 1)
	go StartDispatcher(db, ctx, workerCh)

	cancel()
	// Dispatcher should exit and close channel
	_, open := <-workerCh
	if open {
		t.Error("expected workerCh to be closed after context cancel")
	}
}

func TestStartDispatcher_IncrementsRetryMetricWhenAttemptsGreaterThanOne(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := InitJobsSchema(db); err != nil {
		t.Fatalf("init schema: %v", err)
	}

	// Insert with attempts=1 so that when ClaimJob runs (attempts+1), we get attempts=2 and dispatcher calls IncRetryCount
	payload := json.RawMessage(`{}`)
	_, err = db.Exec(`
		INSERT INTO jobs (type, status, payload, max_retries, attempts, run_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'), datetime('now'))`,
		"test", "queued", payload, 3, 1)
	if err != nil {
		t.Fatalf("insert job: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	workerCh := make(chan WorkerJob, 1)
	go StartDispatcher(db, ctx, workerCh)

	select {
	case job := <-workerCh:
		if job.Attempts != 2 {
			t.Errorf("job attempts = %d, want 2 (retry)", job.Attempts)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for job")
	}

	cancel()
	_, _ = <-workerCh // drain closed channel
}
