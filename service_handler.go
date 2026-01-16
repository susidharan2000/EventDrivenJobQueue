package main

import (
	"database/sql"
	"errors"
)

func executeJob(db *sql.DB, job workerJob) error {
	switch job.Type {
	case "email":
		err := sendMail(job.Payload)
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid type request")
	}
	return nil
}
