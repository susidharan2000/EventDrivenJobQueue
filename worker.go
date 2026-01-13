package main

import (
	"database/sql"
)

func startWorker(db *sql.DB) {
	for job := range workerCh {
		go func() {
			err := executeJob(job)

			if err != nil {
				markJobFailed(db, job)
			} else {
				markJobDone(db, job.Id)
			}
		}()
	}
}
