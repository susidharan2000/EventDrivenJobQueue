package main

import (
	"database/sql"
	"errors"
	"log"
	"math/rand"
	"time"
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

func executeJob(job workerJob) error {
	log.Println("executing job:", job.Id)
	// simulate work
	time.Sleep(1 * time.Second)
	r := rand.Intn(200)
	if r%2 == 0 {
		return errors.New("job execution failed")
	}
	return nil
}
