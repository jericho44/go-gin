package repository

import (
	"context"
	"database/sql"

	"gin-golang-app/internal/database"
)

// Repository defines the base interface for all repositories
type Repository interface {
	// GetDB returns the underlying database connection
	GetDB() *sql.DB
	// BeginTx starts a new transaction
	BeginTx(ctx context.Context) (*sql.Tx, error)
}

// BaseRepository provides common functionality for all repositories
type BaseRepository struct {
	db *database.Database
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *database.Database) *BaseRepository {
	return &BaseRepository{
		db: db,
	}
}

// GetDB returns the underlying database connection
func (r *BaseRepository) GetDB() *sql.DB {
	return r.db.DB
}

// BeginTx starts a new transaction
func (r *BaseRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.DB.BeginTx(ctx, nil)
}
