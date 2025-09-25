package usecase

import (
	"context"

	"github.com/zainokta/item-sync/config"
	pkgErrors "github.com/zainokta/item-sync/internal/errors"
	"github.com/zainokta/item-sync/internal/item/jobs"
	"github.com/zainokta/item-sync/pkg/api"
	"github.com/zainokta/item-sync/pkg/logger"
)

type SyncItemsUseCase struct {
	cfg      *config.Config
	itemRepo ItemRepository
	jobRepo  JobRepository
	logger   logger.Logger
}

func NewSyncItemsUseCase(cfg *config.Config, itemRepo ItemRepository, jobRepo JobRepository, logger logger.Logger) *SyncItemsUseCase {
	return &SyncItemsUseCase{
		cfg:      cfg,
		itemRepo: itemRepo,
		jobRepo:  jobRepo,
		logger:   logger,
	}
}

type SyncItemsRequest struct {
	ForceSync bool                   `json:"force_sync"`
	APISource string                 `json:"api_source"`
	Operation string                 `json:"operation"`
	Params    map[string]interface{} `json:"params"`
}

type SyncItemsResponse struct {
	Errors  []string `json:"errors,omitempty"`
	Status  string   `json:"status"`
	Message string   `json:"message"`
}

func (uc *SyncItemsUseCase) Execute(ctx context.Context, req SyncItemsRequest) (SyncItemsResponse, error) {
	// Create API client based on the requested API source
	apiClient, err := api.NewAPIClient(req.APISource, uc.cfg.API, uc.cfg.Retry, uc.logger)
	if err != nil {
		uc.logger.Error("Failed to create API client", "api_source", req.APISource, "error", err)
		return SyncItemsResponse{}, pkgErrors.ExternalAPIFailed(err)
	}

	// Create sync job instance
	syncJob := jobs.NewSyncJob(
		"manual_sync",
		uc.itemRepo,
		uc.jobRepo,
		apiClient,
		req.APISource,
		uc.logger,
		*uc.cfg,
		req.Params,
	)

	// Start background sync job
	go uc.executeBackgroundSync(context.Background(), syncJob)

	return SyncItemsResponse{
		Errors:  make([]string, 0),
		Status:  "accepted",
		Message: "Sync job has been accepted for background processing",
	}, nil
}

func (uc *SyncItemsUseCase) executeBackgroundSync(ctx context.Context, syncJob *jobs.SyncJob) {
	uc.logger.Info("Starting background sync job", "job_name", syncJob.Name())

	if err := syncJob.Execute(ctx); err != nil {
		uc.logger.Error("Background sync job failed", "job_name", syncJob.Name(), "error", err)
	} else {
		uc.logger.Info("Background sync job completed successfully", "job_name", syncJob.Name())
	}
}
