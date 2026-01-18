package jobqueue

import (
	"database/sql"
	"log"
	"os"
	"strconv"
	"sync"
)

func StartWorkers(db *sql.DB, workerCh chan WorkerJob, wg *sync.WaitGroup) {
	nStr := os.Getenv("WORKER_COUNT")
	n, err := strconv.Atoi(nStr)
	if err != nil || n <= 0 {
		n = 5
	}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go worker(db, workerCh, wg)
	}
}

func worker(db *sql.DB, workerCh chan WorkerJob, wg *sync.WaitGroup) {
	defer wg.Done()
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
