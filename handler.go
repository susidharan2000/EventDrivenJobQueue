package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// Handler
// Producer handler
func CreatejobRequest(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	//request limiter
	select {
	case requestLimiter <- struct{}{}:
		defer func() { <-requestLimiter }()
	default:
		http.Error(w, "server busy, retry later", http.StatusTooManyRequests)
		return
	}

	// method check
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)

	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	jsonData := json.NewDecoder(bytes.NewReader(bodyBytes))
	jsonData.DisallowUnknownFields() //disallow fields
	var req CreateJob
	err = jsonData.Decode(&req)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	err = validateRequestField(&req)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	req.Status = "queued"

	//producer limiter
	producerLimiter <- struct{}{}
	defer func() { <-producerLimiter }()

	err = produceJob(&req, db)
	if err != nil {
		// log.Println(err)
		// ErrorResponse(w, http.StatusBadRequest, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//success response
	SuccessResponse(w, http.StatusCreated, "Job created Successfully")

}

// Validation

func validateRequestField(req *CreateJob) error {
	if req.Type == "" {
		return errors.New("job type field required")
	}
	if len(req.Payload) == 0 {
		return errors.New("payload is required")
	}
	if req.IdempotencyKey != nil && *req.IdempotencyKey == "" {
		return errors.New("IdempotencyKey can't be empty")
	}
	return nil
}

//response Writter

func ErrorResponse(w http.ResponseWriter, s int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(s)
	json.NewEncoder(w).Encode(map[string]string{
		"message": message,
	})
}
func SuccessResponse(w http.ResponseWriter, s int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(s)
	json.NewEncoder(w).Encode(map[string]string{
		"message": message,
	})
}
