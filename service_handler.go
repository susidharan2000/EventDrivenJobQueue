package main

import "errors"

func executeJob(job workerJob) error {
	// log.Println("executing job:", job.Id)
	// // simulate work
	// time.Sleep(1 * time.Second)
	// r := rand.Intn(200)
	// if r%2 == 0 {
	// 	return errors.New("job execution failed")
	// }
	// return nil
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
