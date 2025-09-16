package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zainokta/item-sync/internal/errors"
	"github.com/zainokta/item-sync/internal/item/entity"
	"github.com/zainokta/item-sync/internal/item/usecase"
	"github.com/zainokta/item-sync/pkg/logger"
)

// Ensure the cache implements the required interface
var _ usecase.ItemCache = (*ItemCache)(nil)

type ItemCache struct {
	client *redis.Client
	ttl    time.Duration
	logger logger.Logger
}

func NewItemCache(client *redis.Client, ttl time.Duration, logger logger.Logger) *ItemCache {
	return &ItemCache{
		client: client,
		ttl:    ttl,
		logger: logger,
	}
}

func (c *ItemCache) GetItems(ctx context.Context, key string) ([]entity.Item, error) {
	c.logger.Debug("Cache get items", "key", key)

	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			c.logger.Debug("Cache miss", "key", key)
			return nil, redis.Nil
		}
		c.logger.Error("Cache get items failed", "key", key, "error", err.Error())
		return nil, errors.CacheFailed(err)
	}

	var items []entity.Item
	if err := json.Unmarshal(data, &items); err != nil {
		c.logger.Error("Cache unmarshal failed", "key", key, "error", err.Error())
		return nil, errors.CacheFailed(err)
	}

	c.logger.Debug("Cache hit", "key", key, "items_count", len(items))
	return items, nil
}

func (c *ItemCache) SetItems(ctx context.Context, key string, items []entity.Item, ttl time.Duration) error {
	c.logger.Debug("Cache set items", "key", key, "items_count", len(items), "ttl", ttl)

	data, err := json.Marshal(items)
	if err != nil {
		c.logger.Error("Cache marshal failed", "key", key, "error", err.Error())
		return errors.CacheFailed(err)
	}

	if ttl <= 0 {
		ttl = c.ttl
		c.logger.Debug("Using default TTL", "key", key, "ttl", ttl)
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		c.logger.Error("Cache set failed", "key", key, "error", err.Error())
		return errors.CacheFailed(err)
	}

	c.logger.Debug("Cache set success", "key", key)
	return nil
}

func (c *ItemCache) Invalidate(ctx context.Context, key string) error {
	c.logger.Debug("Cache invalidate", "key", key)

	if err := c.client.Del(ctx, key).Err(); err != nil {
		c.logger.Error("Cache invalidate failed", "key", key, "error", err.Error())
		return errors.CacheFailed(err)
	}

	keys, err := c.client.Keys(ctx, key).Result()
	if err != nil {
		c.logger.Error("Cache get keys failed", "key", key, "error", err.Error())
		return nil
	}

	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			c.logger.Error("Cache delete keys failed", "keys", keys, "error", err.Error())
			return nil
		}
		c.logger.Debug("Cache invalidated multiple keys", "key", key, "deleted_count", len(keys))
	}

	c.logger.Debug("Cache invalidate success", "key", key)
	return nil
}

func (c *ItemCache) GetItem(ctx context.Context, key string) (entity.Item, error) {
	c.logger.Debug("Cache get item", "key", key)

	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			c.logger.Debug("Cache miss", "key", key)
			return entity.Item{}, redis.Nil
		}
		c.logger.Error("Cache get item failed", "key", key, "error", err.Error())
		return entity.Item{}, errors.CacheFailed(err)
	}

	var item entity.Item
	if err := json.Unmarshal(data, &item); err != nil {
		c.logger.Error("Cache unmarshal item failed", "key", key, "error", err.Error())
		return entity.Item{}, errors.CacheFailed(err)
	}

	c.logger.Debug("Cache hit", "key", key, "item_id", item.ID)
	return item, nil
}

func (c *ItemCache) SetItem(ctx context.Context, key string, item entity.Item, ttl time.Duration) error {
	c.logger.Debug("Cache set item", "key", key, "item_id", item.ID, "ttl", ttl)

	data, err := json.Marshal(item)
	if err != nil {
		c.logger.Error("Cache marshal item failed", "key", key, "item_id", item.ID, "error", err.Error())
		return errors.CacheFailed(err)
	}

	if ttl <= 0 {
		ttl = c.ttl
		c.logger.Debug("Using default TTL", "key", key, "ttl", ttl)
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		c.logger.Error("Cache set item failed", "key", key, "item_id", item.ID, "error", err.Error())
		return errors.CacheFailed(err)
	}

	c.logger.Debug("Cache set item success", "key", key, "item_id", item.ID)
	return nil
}
