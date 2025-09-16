package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zainokta/item-sync/internal/errors"
	"github.com/zainokta/item-sync/internal/item/entity"
	"github.com/zainokta/item-sync/internal/item/usecase"
	"github.com/zainokta/item-sync/pkg/logger"
)

// Ensure the repository implements the required interfaces
var _ usecase.ItemRepository = (*ItemRepository)(nil)

type ItemRepository struct {
	db     *sql.DB
	logger logger.Logger
}

func NewItemRepository(db *sql.DB, logger logger.Logger) *ItemRepository {
	return &ItemRepository{
		db:     db,
		logger: logger,
	}
}

func (r *ItemRepository) Save(ctx context.Context, item entity.Item) error {
	r.logger.Debug("Repository save item", "external_id", item.ExternalID, "api_source", item.APISource)

	query := `
		INSERT INTO items (title, description, external_id, api_source, extend_info, last_synced_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	var extendInfoJSON []byte
	if item.ExtendInfo != nil {
		var err error
		extendInfoJSON, err = json.Marshal(item.ExtendInfo)
		if err != nil {
			r.logger.Error("Repository marshal extend_info failed", "external_id", item.ExternalID, "error", err.Error())
			return errors.DatabaseError(err)
		}
	}

	result, err := r.db.ExecContext(ctx, query,
		item.Title, item.Description, item.ExternalID, item.APISource,
		string(extendInfoJSON), item.SyncedAt, item.CreatedAt, item.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Repository save failed", "external_id", item.ExternalID, "error", err.Error())
		return errors.DatabaseError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		r.logger.Error("Repository get last insert ID failed", "external_id", item.ExternalID, "error", err.Error())
		return errors.DatabaseError(err)
	}

	r.logger.Info("Repository save success", "id", id, "external_id", item.ExternalID, "api_source", item.APISource)
	return nil
}

func (r *ItemRepository) FindByID(ctx context.Context, id int) (entity.Item, error) {
	r.logger.Debug("Repository find by ID", "id", id)

	query := `
		SELECT id, title, description, external_id, api_source, extend_info, last_synced_at, created_at, updated_at
		FROM items 
		WHERE id = ?
	`

	var item entity.Item
	var extendInfoJSON string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID, &item.Title, &item.Description, &item.ExternalID, &item.APISource,
		&extendInfoJSON, &item.SyncedAt, &item.CreatedAt, &item.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Debug("Repository item not found", "id", id)
			return entity.Item{}, errors.ItemNotFound()
		}
		r.logger.Error("Repository find by ID failed", "id", id, "error", err.Error())
		return entity.Item{}, errors.DatabaseError(err)
	}

	if extendInfoJSON != "" {
		if err := json.Unmarshal([]byte(extendInfoJSON), &item.ExtendInfo); err != nil {
			r.logger.Error("Repository unmarshal extend_info failed", "id", id, "error", err.Error())
			return entity.Item{}, errors.DatabaseError(err)
		}
	}

	r.logger.Debug("Repository find by ID success", "id", id, "external_id", item.ExternalID)
	return item, nil
}

func (r *ItemRepository) FindAll(ctx context.Context, limit, offset int) ([]entity.Item, error) {
	r.logger.Debug("Repository find all", "limit", limit, "offset", offset)

	query := `
		SELECT id, title, description, external_id, api_source, extend_info, last_synced_at, created_at, updated_at
		FROM items 
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		r.logger.Error("Repository find all failed", "limit", limit, "offset", offset, "error", err.Error())
		return nil, errors.DatabaseError(err)
	}
	defer rows.Close()

	var items []entity.Item
	for rows.Next() {
		var item entity.Item
		var extendInfoJSON string

		err := rows.Scan(
			&item.ID, &item.Title, &item.Description, &item.ExternalID, &item.APISource,
			&extendInfoJSON, &item.SyncedAt, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Repository scan item failed", "error", err.Error())
			return nil, errors.DatabaseError(err)
		}

		if extendInfoJSON != "" {
			if err := json.Unmarshal([]byte(extendInfoJSON), &item.ExtendInfo); err != nil {
				r.logger.Error("Repository unmarshal extend_info failed", "id", item.ID, "error", err.Error())
				return nil, errors.DatabaseError(err)
			}
		}

		items = append(items, item)
	}

	r.logger.Debug("Repository find all success", "count", len(items))
	return items, nil
}

func (r *ItemRepository) FindByAPISource(ctx context.Context, apiSource string, limit, offset int) ([]entity.Item, error) {
	r.logger.Debug("Repository find by API source", "api_source", apiSource, "limit", limit, "offset", offset)

	query := `
		SELECT id, title, description, external_id, api_source, extend_info, last_synced_at, created_at, updated_at
		FROM items 
		WHERE api_source = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, apiSource, limit, offset)
	if err != nil {
		r.logger.Error("Repository find by API source failed", "api_source", apiSource, "error", err.Error())
		return nil, errors.DatabaseError(err)
	}
	defer rows.Close()

	var items []entity.Item
	for rows.Next() {
		var item entity.Item
		var extendInfoJSON string

		err := rows.Scan(
			&item.ID, &item.Title, &item.Description, &item.ExternalID, &item.APISource,
			&extendInfoJSON, &item.SyncedAt, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Repository scan item failed", "error", err.Error())
			return nil, errors.DatabaseError(err)
		}

		if extendInfoJSON != "" {
			if err := json.Unmarshal([]byte(extendInfoJSON), &item.ExtendInfo); err != nil {
				r.logger.Error("Repository unmarshal extend_info failed", "id", item.ID, "error", err.Error())
				return nil, errors.DatabaseError(err)
			}
		}

		items = append(items, item)
	}

	r.logger.Debug("Repository find by API source success", "api_source", apiSource, "count", len(items))
	return items, nil
}

func (r *ItemRepository) FindByStatus(ctx context.Context, status string, limit, offset int) ([]entity.Item, error) {
	r.logger.Debug("Repository find by status", "status", status, "limit", limit, "offset", offset)

	query := `
		SELECT id, title, description, external_id, api_source, extend_info, last_synced_at, created_at, updated_at
		FROM items 
		WHERE JSON_EXTRACT(extend_info, '$.status') = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, status, limit, offset)
	if err != nil {
		r.logger.Error("Repository find by status failed", "status", status, "error", err.Error())
		return nil, errors.DatabaseError(err)
	}
	defer rows.Close()

	var items []entity.Item
	for rows.Next() {
		var item entity.Item
		var extendInfoJSON string

		err := rows.Scan(
			&item.ID, &item.Title, &item.Description, &item.ExternalID, &item.APISource,
			&extendInfoJSON, &item.SyncedAt, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Repository scan item failed", "error", err.Error())
			return nil, errors.DatabaseError(err)
		}

		if extendInfoJSON != "" {
			if err := json.Unmarshal([]byte(extendInfoJSON), &item.ExtendInfo); err != nil {
				r.logger.Error("Repository unmarshal extend_info failed", "id", item.ID, "error", err.Error())
				return nil, errors.DatabaseError(err)
			}
		}

		items = append(items, item)
	}

	r.logger.Debug("Repository find by status success", "status", status, "count", len(items))
	return items, nil
}

func (r *ItemRepository) FindByType(ctx context.Context, itemType string, limit, offset int) ([]entity.Item, error) {
	r.logger.Debug("Repository find by type (deprecated, using API source)", "item_type", itemType)
	return r.FindByAPISource(ctx, itemType, limit, offset)
}

func (r *ItemRepository) UpsertWithHash(ctx context.Context, apiSource string, externalItem entity.ExternalItem) error {
	now := time.Now()

	extendInfoJSON, err := json.Marshal(externalItem.ExtendInfo)
	if err != nil {
		r.logger.Error("Repository marshal extend_info failed", "external_id", externalItem.ID, "error", err.Error())
		return errors.DatabaseError(err)
	}

	contentHash := r.calculateContentHash(externalItem.Title, string(extendInfoJSON))

	query := `
		INSERT INTO items (title, description, external_id, api_source, extend_info, content_hash, last_synced_at, created_at, updated_at, sync_attempts)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1)
		ON DUPLICATE KEY UPDATE
			title = VALUES(title),
			description = VALUES(description),
			extend_info = CASE 
				WHEN content_hash != VALUES(content_hash) THEN VALUES(extend_info)
				ELSE extend_info
			END,
			content_hash = VALUES(content_hash),
			last_synced_at = VALUES(last_synced_at),
			updated_at = VALUES(updated_at),
			sync_attempts = sync_attempts + 1,
			last_sync_error = NULL
	`

	_, err = r.db.ExecContext(ctx, query,
		externalItem.Title,
		"", // description - might be extracted from extend_info if needed
		externalItem.ID,
		apiSource,
		string(extendInfoJSON),
		contentHash,
		now,
		now,
		now,
	)

	if err != nil {
		r.logger.Error("Repository upsert with hash failed", "external_id", externalItem.ID, "api_source", apiSource, "error", err.Error())
		return errors.DatabaseError(err)
	}

	r.logger.Debug("Repository upsert with hash success", "external_id", externalItem.ID, "api_source", apiSource, "content_hash", contentHash)
	return nil
}

func (r *ItemRepository) calculateContentHash(title string, extendInfoJSON string) string {
	content := fmt.Sprintf("%s:%s", title, extendInfoJSON)
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)
}
