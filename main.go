package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)

// producer Limiter
// The "producerLimiter" is a channel used for limiting the number of concurrent producers in the
// system.
var requestLimiter = make(chan struct{}, 100)
var producerLimiter = make(chan struct{}, 50)

var workerCh = make(chan workerJob, 10)

func main() {

	db, err := sql.Open("sqlite", "jobs.db")
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	//Configurations
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec("PRAGMA synchronous=NORMAL;"); err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout=5000;"); err != nil {
		log.Fatal(err)
	}

	// Inilize Schema
	if err = InitJobsSchema(db); err != nil {
		log.Fatal(err)
	}

	defer db.Close()
	// set maxConnection
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	//start
	go startDispatcher(db)
	go startWorker(db)

	router := NewRouter(db)
	port := 8080
	adr := fmt.Sprintf(":%v", port)

	log.Fatal(http.ListenAndServe(adr, router))
}
