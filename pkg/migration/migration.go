package migration

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/zainokta/item-sync/pkg/logger"
	"database/sql"
)

type Migrator struct {
	migrate *migrate.Migrate
	logger  logger.Logger
}

type Config struct {
	DatabaseURL    string
	MigrationsPath string
	Logger         logger.Logger
}

// NewMigrator creates a new migration instance
func NewMigrator(config Config) (*Migrator, error) {
	if config.DatabaseURL == "" {
		return nil, errors.New("database URL is required")
	}
	if config.MigrationsPath == "" {
		return nil, errors.New("migrations path is required")
	}

	// Create database connection for migrate
	db, err := sql.Open("mysql", config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Create MySQL driver instance
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create MySQL driver: %w", err)
	}

	// Create migrate instance with file source
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", config.MigrationsPath),
		"mysql",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return &Migrator{
		migrate: m,
		logger:  config.Logger,
	}, nil
}

// Up runs all pending migrations
func (m *Migrator) Up() error {
	m.logger.Info("Running database migrations up...")
	
	err := m.migrate.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			m.logger.Info("No pending migrations found")
			return nil
		}
		return fmt.Errorf("failed to run migrations up: %w", err)
	}
	
	m.logger.Info("Successfully applied all pending migrations")
	return nil
}

// Down rolls back one migration
func (m *Migrator) Down() error {
	m.logger.Info("Rolling back one migration...")
	
	err := m.migrate.Steps(-1)
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			m.logger.Info("No migrations to rollback")
			return nil
		}
		return fmt.Errorf("failed to rollback migration: %w", err)
	}
	
	m.logger.Info("Successfully rolled back one migration")
	return nil
}

// Version returns current migration version
func (m *Migrator) Version() (uint, bool, error) {
	version, dirty, err := m.migrate.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			return 0, false, nil // No migrations applied yet
		}
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}
	return version, dirty, nil
}

// Drop drops the entire database
func (m *Migrator) Drop() error {
	m.logger.Warn("Dropping entire database schema...")
	
	err := m.migrate.Drop()
	if err != nil {
		return fmt.Errorf("failed to drop database: %w", err)
	}
	
	m.logger.Info("Successfully dropped database schema")
	return nil
}

// Force sets the migration version without running migrations (for recovery)
func (m *Migrator) Force(version int) error {
	m.logger.Warn("Forcing migration version", "version", version)
	
	err := m.migrate.Force(version)
	if err != nil {
		return fmt.Errorf("failed to force migration version: %w", err)
	}
	
	m.logger.Info("Successfully forced migration version", "version", version)
	return nil
}

// Close closes the migration instance
func (m *Migrator) Close() error {
	sourceErr, databaseErr := m.migrate.Close()
	if sourceErr != nil {
		return fmt.Errorf("failed to close migration source: %w", sourceErr)
	}
	if databaseErr != nil {
		return fmt.Errorf("failed to close database connection: %w", databaseErr)
	}
	return nil
}