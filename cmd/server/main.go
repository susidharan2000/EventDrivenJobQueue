package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/susi/EventDrivenJobQueue/internal/jobqueue"
	_ "modernc.org/sqlite"
)

// request limiter
var requestLimiter = make(chan struct{}, 100)

// producer Limiter
var producerLimiter = make(chan struct{}, 50)

var workerCh = make(chan jobqueue.WorkerJob, 10)

var wg = sync.WaitGroup{}

func main() {

	//shutdown signal catch
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()
	go func() {
		<-ctx.Done()
		fmt.Println("shutdown started")
	}()

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
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		log.Fatal(err)
	}

	//load .env
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file")
	}

	// Inilize Schema
	if err = jobqueue.InitJobsSchema(db); err != nil {
		log.Fatal(err)
	}

	defer db.Close()
	// set maxConnection
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	//start
	go jobqueue.StartDispatcher(db, ctx, workerCh)
	go jobqueue.StartWorkers(db, workerCh, &wg)
	go jobqueue.StartVisibilityReaper(db)

	router := jobqueue.NewRouter(db, requestLimiter, producerLimiter)
	port := 8080
	adr := fmt.Sprintf(":%v", port)
	go func() {
		err := http.ListenAndServe(adr, router)
		if err != nil {
			log.Println("http server error:", err)
		}
	}()

	<-ctx.Done()
	close(workerCh) //close the dispatcher and worker communication
	wg.Wait()       //wait untill all the workers conplete before shutdown

	//time.Sleep(5 * time.Second)
}
