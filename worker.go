package main

import (
	"database/sql"
	"os"
	"strconv"
)

func startWorkers(db *sql.DB) {
	nStr := os.Getenv("WORKER_COUNT")
	n, err := strconv.Atoi(nStr)
	if err != nil || n <= 0 {
		n = 5
	}
	for i := 0; i < n; i++ {
		go worker(db)
	}
}

func worker(db *sql.DB) {
	for job := range workerCh {
		err := executeJob(job)

		if err != nil {
			markJobFailed(db, job)
		} else {
			markJobDone(db, job.Id)
		}
	}
}
