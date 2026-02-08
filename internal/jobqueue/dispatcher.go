package jobqueue

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/susi/EventDrivenJobQueue/internal/metrics"
)

func StartDispatcher(db *sql.DB, ctx context.Context, workerCh chan WorkerJob) {
	defer close(workerCh) //close the dispatcher and worker communication
	// pull the job from the db and assign to the worker
	for {
		select {
		case <-ctx.Done():
			log.Println("Dispatcher is Dead")
			return
		default:
		}
		job, err := ClaimJob(db, ctx)
		if job.Attempts > 1 {
			metrics.M.IncRetryCount()
		}
		if err == sql.ErrNoRows {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if err != nil {
			log.Println("dispatcher error:", err)
			time.Sleep(time.Second)
			continue
		}
		//metrics
		metrics.M.DecQueueDepth()

		queueTimeMs := job.StartedAt.Sub(job.RunAt).Milliseconds()
		metrics.M.AddQueueTime(queueTimeMs)

		select {
		case workerCh <- job:
		case <-ctx.Done():
			log.Println("Dispatcher is Dead")
			return
		}
	}
}
