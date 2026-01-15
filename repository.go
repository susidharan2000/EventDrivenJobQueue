package main

import (
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

	// Side Effect Idempotency Table
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS side_effects (
		job_id INTEGER PRIMARY KEY,
		effect_type TEXT NOT NULL,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (job_id) REFERENCES jobs(id)
	);
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
func ClaimJob(db *sql.DB) (workerJob, error) {
	row := db.QueryRow(`UPDATE jobs SET status = 'processing',started_at = datetime('now') WHERE id = (SELECT id FROM jobs WHERE status = 'queued' AND run_at <= datetime('now') ORDER BY run_at LIMIT 1)
	RETURNING id, type, status, payload, max_retries, attempts,run_at`)
	var job workerJob
	err := row.Scan(&job.Id, &job.Type, &job.Status, &job.Payload, &job.MaxRetries, &job.Attempts, &job.RunAt)
	if err == sql.ErrNoRows {
		return workerJob{}, err
	}

	if err != nil {
		return workerJob{}, err
	}
	return job, nil
}

// mark the job Failed //retry and back off
func markJobFailed(db *sql.DB, job workerJob) {
	var att int
	cont MaxRetries=5
	if job.Attempts < job.MaxRetries {
		att = job.Attempts + 1
		for i:=1;i<=MaxRetries;i++ {
			res, err := db.Exec(`UPDATE jobs SET status = 'queued',run_at = datetime('now', '+10 seconds'),attempts = ? WHERE id = ? AND status='processing'`, att, job.Id)
			if err != nil {
				if isLockedError(err) {
					time.Sleep(1000 * time.Millisecond)
					continue
				}
				fmt.Printf("Failed state update failed")
				return
			}
			affected, err := res.RowsAffected()
			if err != nil {
				log.Println("Rows Affected error:", err)
				return
			}
			if affected == 0 {
				fmt.Printf("Failed state update failed")
				return
			}
			fmt.Printf("Attempts:%v, MaxRetries:%v \n", att, job.MaxRetries)
			return
		}
	} else {
		for i:=1;i<=MaxRetries;i++ {
			res, err := db.Exec(`UPDATE jobs SET status = 'failed',finished_at = datetime('now') WHERE id = ? AND status='processing'`, job.Id)
			if err != nil {
				if isLockedError(err) {
					time.Sleep(time.Millisecond)
					continue
				}
				log.Println("markJobFailed error:", err)
				return
			}
			affected, err := res.RowsAffected()
			if err != nil {
				log.Println("Rows Affected error:", err)
				return
			}
			if affected == 0 {
				fmt.Printf("Failed state update failed")
				return
			}
			log.Println("Job Marked", err)
			return
		}
	}
}

func markJobDone(db *sql.DB, id int) {
	cont MaxRetries=5
	for i:=1;i<=MaxRetries;i++ {
		res, err := db.Exec(`UPDATE jobs SET status = 'done',finished_at = datetime('now') WHERE id = ? AND status='processing'`, id)

		if err != nil {
			if isLockedError(err) {
				time.Sleep(1000 * time.Millisecond)
				continue
			}
			log.Println(err)
			return
		}

		affected, err := res.RowsAffected()
		if err != nil {
			log.Println("RowsAffected error:", err)
			return
		}
		if affected == 0 {
			log.Println("mark Jobs success error:", err)
			return
		}

		fmt.Printf("%v: Job Done\n", id)
		return
	}
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

// apply Side-Effect
func applySideEffect(db *sql.DB, job workerJob) error {
	_, err = db.Exec(`INSERT INTO side_effects (job_id,effect_type,applied_at) VALUES (?,?,datetime('now'));`, job.Id, job.Type)
	if err != nil {
		return err
	}
	return nil
}
