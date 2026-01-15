package main

import (
	"database/sql"
	"log"
	"time"
)

func startVisibilityReaper(db *sql.DB) {
	timer := time.NewTicker(30 * time.Second)
	go func() {
		for range timer.C {
			_, err := db.Exec(`UPDATE jobs SET status = 'queued',attempts = attempts+1,run_at = datetime('now') WHERE status = 'processing' AND started_at < datetime('now','-1 minutes') AND attempts < max_retries`)
			if err != nil {
				log.Println("visibility reaper error:", err)
			}
		}
	}()
}
