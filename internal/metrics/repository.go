package metrics

import "database/sql"

// count pending jobs
func CountPendingJobs(db *sql.DB) (int64, error) {
	var count int64
	if err := db.QueryRow(`
	SELECT COUNT(*)
	FROM jobs
	WHERE status = 'queued';
    `).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// count Successful jobs
func CountCompletedJobs(db *sql.DB) (int64, error) {
	var count int64
	if err := db.QueryRow(`
	SELECT COUNT(*)
	FROM jobs
	WHERE status = 'done';
    `).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

//count failed jobs

func CountFailedJobs(db *sql.DB) (int64, error) {
	var count int64
	if err := db.QueryRow(`
	SELECT COUNT(*)
	FROM jobs
	WHERE status = 'failed';
    `).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// count total retries
func CountTotalRetries(db *sql.DB) (int64, error) {
	var total int64

	err := db.QueryRow(`
        SELECT COALESCE(SUM(attempts - 1), 0)
        FROM jobs
        WHERE attempts > 1
    `).Scan(&total)

	if err != nil {
		return 0, err
	}

	return total, nil
}
