package repository

import (
	"context"
	"database/sql"
	"encoding/json"

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
		INSERT INTO items (title, description, external_id, api_source, extend_info, synced_at, created_at, updated_at)
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

func (r *ItemRepository) Update(ctx context.Context, item entity.Item) error {
	r.logger.Debug("Repository update item", "id", item.ID, "external_id", item.ExternalID, "api_source", item.APISource)
	
	query := `
		UPDATE items 
		SET title = ?, description = ?, external_id = ?, api_source = ?, extend_info = ?, 
		    synced_at = ?, updated_at = ?
		WHERE id = ?
	`

	var extendInfoJSON []byte
	if item.ExtendInfo != nil {
		var err error
		extendInfoJSON, err = json.Marshal(item.ExtendInfo)
		if err != nil {
			r.logger.Error("Repository marshal extend_info failed", "id", item.ID, "error", err.Error())
			return errors.DatabaseError(err)
		}
	}

	result, err := r.db.ExecContext(ctx, query,
		item.Title, item.Description, item.ExternalID, item.APISource,
		string(extendInfoJSON), item.SyncedAt, item.UpdatedAt, item.ID,
	)

	if err != nil {
		r.logger.Error("Repository update failed", "id", item.ID, "error", err.Error())
		return errors.DatabaseError(err)
	}

	rowsAffected, _ := result.RowsAffected()
	r.logger.Info("Repository update success", "id", item.ID, "rows_affected", rowsAffected)
	return nil
}

func (r *ItemRepository) FindByID(ctx context.Context, id int) (entity.Item, error) {
	r.logger.Debug("Repository find by ID", "id", id)
	
	query := `
		SELECT id, title, description, external_id, api_source, extend_info, synced_at, created_at, updated_at
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

func (r *ItemRepository) FindByExternalID(ctx context.Context, externalID int) (entity.Item, error) {
	r.logger.Debug("Repository find by external ID", "external_id", externalID)
	
	query := `
		SELECT id, title, description, external_id, api_source, extend_info, synced_at, created_at, updated_at
		FROM items 
		WHERE external_id = ?
	`

	var item entity.Item
	var extendInfoJSON string

	err := r.db.QueryRowContext(ctx, query, externalID).Scan(
		&item.ID, &item.Title, &item.Description, &item.ExternalID, &item.APISource,
		&extendInfoJSON, &item.SyncedAt, &item.CreatedAt, &item.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Debug("Repository item not found by external ID", "external_id", externalID)
			return entity.Item{}, errors.ItemNotFound()
		}
		r.logger.Error("Repository find by external ID failed", "external_id", externalID, "error", err.Error())
		return entity.Item{}, errors.DatabaseError(err)
	}

	if extendInfoJSON != "" {
		if err := json.Unmarshal([]byte(extendInfoJSON), &item.ExtendInfo); err != nil {
			r.logger.Error("Repository unmarshal extend_info failed", "external_id", externalID, "error", err.Error())
			return entity.Item{}, errors.DatabaseError(err)
		}
	}

	r.logger.Debug("Repository find by external ID success", "id", item.ID, "external_id", externalID)
	return item, nil
}

func (r *ItemRepository) FindAll(ctx context.Context, limit, offset int) ([]entity.Item, error) {
	r.logger.Debug("Repository find all", "limit", limit, "offset", offset)
	
	query := `
		SELECT id, title, description, external_id, api_source, extend_info, synced_at, created_at, updated_at
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
		SELECT id, title, description, external_id, api_source, extend_info, synced_at, created_at, updated_at
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
		SELECT id, title, description, external_id, api_source, extend_info, synced_at, created_at, updated_at
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
