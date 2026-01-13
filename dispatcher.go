package main

import (
	"database/sql"
	"log"
	"time"
)

func startDispatcher(db *sql.DB) {
	// pull the job from the db and assign to the worker
	for {
		job, err := ClaimJob(db)
		if err == sql.ErrNoRows {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if err != nil {
			log.Println("dispatcher error:", err)
			time.Sleep(time.Second)
			continue
		}
		workerCh <- job
	}

}
