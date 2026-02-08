package jobqueue

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

// Inilize Schema
func InitJobsSchema(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS jobs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT NOT NULL,
		status TEXT NOT NULL CHECK (
			status IN ('queued', 'processing', 'done', 'failed')
		),
		payload BLOB NOT NULL,
		attempts INTEGER NOT NULL DEFAULT 0,
		max_retries INTEGER NOT NULL,
		run_at DATETIME NOT NULL,
		idempotency_key TEXT UNIQUE,
		started_at DATETIME,
		finished_at DATETIME,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)
	`)
	if err != nil {
		return err
	}

	// for Dispatcher lookup
	if _, err := db.Exec(`
	CREATE INDEX IF NOT EXISTS idx_jobs_status_run_at
	ON jobs (status, run_at);
	`); err != nil {
		return err
	}
	return nil
}

// job - Producer
func produceJob(req *CreateJob, db *sql.DB) error {
	_, err := db.Exec(`
	INSERT INTO jobs (type,status,payload,max_retries,run_at,idempotency_key) VALUES (?,?,?,?,datetime('now'),?)`,
		req.Type, req.Status, req.Payload, req.MaxRetries, req.IdempotencyKey)
	if err != nil {
		log.Printf("INSERT FAILED: %v", err)
		return err
	}
	return nil
}

// Claim Job
func ClaimJob(db *sql.DB, ctx context.Context) (WorkerJob, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return WorkerJob{}, err
	}
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	row := tx.QueryRowContext(ctx, `UPDATE jobs SET status = 'processing',attempts = attempts+1,started_at = datetime('now') WHERE id = (SELECT id FROM jobs WHERE status = 'queued' AND run_at <= datetime('now') ORDER BY run_at LIMIT 1)
	RETURNING id, type, status, payload, max_retries, attempts,run_at,started_at,finished_at,created_at,updated_at`)
	var job WorkerJob
	err = row.Scan(&job.Id, &job.Type, &job.Status, &job.Payload, &job.MaxRetries, &job.Attempts, &job.RunAt, &job.StartedAt, &job.FinishedAt, &job.CreatedAt, &job.UpdatedAt)
	if err == sql.ErrNoRows {
		return WorkerJob{}, err
	}

	if err != nil {
		return WorkerJob{}, err
	}
	if err := tx.Commit(); err != nil {
		return WorkerJob{}, err
	}
	committed = true
	return job, nil
}

// mark the job Failed //retry and back off
func markJobFailed(db *sql.DB, job WorkerJob) error {
	const dbRetryAttempts = 5

	for i := 0; i < dbRetryAttempts; i++ {
		res, err := db.Exec(`UPDATE jobs SET status = 'failed',finished_at = datetime('now') WHERE id = ? AND status='processing'`, job.Id)
		if err != nil {
			if isLockedError(err) {
				time.Sleep(200 * time.Millisecond)
				continue
			}
			return err
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			return nil
		}
		log.Printf("Job %s marked FAILED", job.Id)
		return nil
	}
	return fmt.Errorf("failed to mark job failed after retries")
}

func markJobDone(db *sql.DB, id int) error {
	const dbRetryAttempts = 5
	for i := 1; i <= dbRetryAttempts; i++ {
		res, err := db.Exec(`UPDATE jobs SET status = 'done',finished_at = datetime('now') WHERE id = ? AND status='processing'`, id)

		if err != nil {
			if isLockedError(err) {
				time.Sleep(200 * time.Millisecond)
				continue
			}
			log.Println(err)
			return err
		}

		affected, _ := res.RowsAffected()
		if affected == 0 {
			log.Printf("job %d update affected 0 rows", id)
			return nil
		}

		log.Printf("job %d marked done", id)
		return nil
	}
	return fmt.Errorf("failed to mark job done after retries")
}

func requeueJob(db *sql.DB, job WorkerJob) error {
	_, err := db.Exec(`
		UPDATE jobs
		SET
			status = 'queued',
			started_at = NULL,
			run_at = datetime('now', '+30 seconds'),
			updated_at = datetime('now')
		WHERE id = ?
		AND status = 'processing'
	`, job.Id)
	if err != nil {
		return err
	}
	return nil
}

// is database lock check
func isLockedError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "database is locked") ||
		strings.Contains(msg, "database is busy")
}
