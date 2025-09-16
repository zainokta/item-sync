package jobs

import (
	"context"
	"time"
)

type JobRepository interface {
	CreateSyncJobRecord(ctx context.Context, name string, apiType string) (int64, error)
	UpdateSyncJobRecord(ctx context.Context, jobID int64, status string, processed, succeeded, failed int, lastErr error, executionTime time.Duration) error
}
