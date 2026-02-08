package jobqueue

import (
	"database/sql"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/susi/EventDrivenJobQueue/internal/metrics"
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
		metrics.M.IncActiveWorkers()
		start := time.Now()
		err := executeJob(db, job)
		if err != nil {
			if strings.Contains(err.Error(), "Daily user sending limit exceeded") {
				if err := markJobFailed(db, job); err == nil {
					metrics.M.IncFailureCount()
					metrics.M.IncDeadLetterCount()
				} else {
					log.Println(err)
				}
			} else if job.Attempts >= job.MaxRetries {
				if err := markJobFailed(db, job); err == nil {
					metrics.M.IncFailureCount()
					metrics.M.IncDeadLetterCount()
				} else {
					log.Println(err)
				}
			} else {
				if err := requeueJob(db, job); err != nil {
					log.Println(err)
				}
			}
		} else {
			if err := markJobDone(db, job.Id); err == nil {
				metrics.M.IncSuccessCount()
			} else {
				log.Println("markJobDone error:", err)
			}
		}
		//mertics
		ExecutionTimeMs := time.Since(start).Milliseconds()
		metrics.M.AddExecutionTime(ExecutionTimeMs)

		metrics.M.DecActiveWorkers()
	}
}
