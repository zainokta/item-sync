package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/zainokta/item-sync/pkg/logger"
)

type JobRepository struct {
	db     *sql.DB
	logger logger.Logger
}

func NewJobRepository(db *sql.DB, logger logger.Logger) *JobRepository {
	return &JobRepository{
		db:     db,
		logger: logger,
	}
}

func (j *JobRepository) CreateSyncJobRecord(ctx context.Context, name string, apiType string) (int64, error) {
	query := `
		INSERT INTO sync_jobs (job_name, api_source, status, started_at) 
		VALUES (?, ?, 'running', ?)
	`

	result, err := j.db.ExecContext(ctx, query, name, apiType, time.Now())
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (j *JobRepository) UpdateSyncJobRecord(ctx context.Context, jobID int64, status string, processed, succeeded, failed int, lastErr error, executionTime time.Duration) error {
	var errorMessage sql.NullString
	if lastErr != nil {
		errorMessage = sql.NullString{String: lastErr.Error(), Valid: true}
	}

	query := `
		UPDATE sync_jobs 
		SET status = ?, completed_at = ?, items_processed = ?, items_succeeded = ?, 
		    items_failed = ?, error_message = ?, execution_time_ms = ?
		WHERE id = ?
	`

	_, err := j.db.ExecContext(ctx, query,
		status, time.Now(), processed, succeeded, failed,
		errorMessage, executionTime.Milliseconds(), jobID)

	return err
}
