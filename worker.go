package main

import (
	"database/sql"
	"log"
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

		err := executeJob(db, job)
		if err != nil {
			log.Printf("Execute Job Error: %s", err)
			return
		}
		if err != nil {
			markJobFailed(db, job)
		} else {
			markJobDone(db, job.Id)
		}
	}
}
