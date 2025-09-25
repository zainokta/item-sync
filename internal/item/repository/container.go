package repository

import (
	"database/sql"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zainokta/item-sync/internal/item/usecase"
	"github.com/zainokta/item-sync/pkg/logger"
)

type RepositoryContainer struct {
	ItemRepository usecase.ItemRepository
	JobRepository  usecase.JobRepository
	ItemCache      usecase.ItemCache
}

func NewRepositoryContainer(db *sql.DB, redis *redis.Client, cacheTTL time.Duration, logger logger.Logger) *RepositoryContainer {
	return &RepositoryContainer{
		ItemRepository: NewItemRepository(db, logger),
		JobRepository:  NewJobRepository(db, logger),
		ItemCache:      NewItemCache(redis, cacheTTL, logger),
	}
}

func (c *RepositoryContainer) GetItemRepository() usecase.ItemRepository {
	return c.ItemRepository
}

func (c *RepositoryContainer) GetJobRepository() usecase.JobRepository {
	return c.JobRepository
}

func (c *RepositoryContainer) GetItemCache() usecase.ItemCache {
	return c.ItemCache
}