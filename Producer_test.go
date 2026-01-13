package main

import (
	"bytes"
	"database/sql"
	"net/http"
	"sync"
	"testing"
)

//var clientLimiter = make(chan struct{}, 1000)

func TestConcurrentJobCreation(t *testing.T) {
	db, err := sql.Open("sqlite", "jobs.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	_, _ = db.Exec("PRAGMA journal_mode=WAL;")
	_, _ = db.Exec("PRAGMA synchronous=NORMAL;")
	_, _ = db.Exec("PRAGMA busy_timeout=5000;")

	_, err = db.Exec("DELETE FROM jobs;")
	if err != nil {
		t.Fatal(err)
	}

	requests := 10000
	wg := sync.WaitGroup{}
	wg.Add(requests)

	body := `{
		"type":"email",
		"payload":{
			"email":"susi@gmail.com",
			"subject":"test-subject",
			"message":"hello World!"
		},
		"max_retries": 3
	}`

	for i := 1; i <= requests; i++ {
		go func() {
			defer wg.Done()
			//client limiter
			// clientLimiter <- struct{}{}
			// defer func() { <-clientLimiter }()

			resp, err := http.Post(
				"http://localhost:8080/createJob",
				"application/json",
				bytes.NewReader([]byte(body)),
			)
			if err != nil {
				t.Errorf("request failed: %v", err)
				return
			}
			resp.Body.Close()
		}()
	}
	wg.Wait()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM jobs").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != requests {
		t.Fatalf("expected %d jobs, got %d", requests, count)
	}

}

// req := CreateJob{
// 	Type:   "email",
// 	Status: "queued",
// 	Payload: json.RawMessage(`{
// 		"email":"susi@gmail.com",
// 		"subject":"test-subject",
// 		"message":"hello World!"
// 	}`),
// 	MaxRetries: 3,
// 	RunAt:      time.Now(),
// }
// err := produceJob(&req, db)
// if err != nil {
// 	t.Errorf("insert failed: %v", err)
// 	// ErrorResponse(w, http.StatusBadRequest, err.Error())
// 	//http.Error(w, err.Error(), http.StatusInternalServerError)
// 	return
// }
