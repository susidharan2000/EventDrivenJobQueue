package jobqueue

import (
	"database/sql"
	"time"
	//"errors"
)

// func executeJob(db *sql.DB, job WorkerJob) error {
// 	switch job.Type {
// 	case "email":
// 		err := SendMail(job.Payload)
// 		if err != nil {
// 			return err
// 		}
// 	default:
// 		return errors.New("invalid type request")
// 	}
// 	return nil
// }

func executeJob(db *sql.DB, job WorkerJob) error {
	time.Sleep(100 * time.Millisecond) // simulate real work
	return nil
}
