package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gin-golang-app/internal/database"
	"gin-golang-app/internal/models"
)

// UserRepository defines the interface for user data access operations
type UserRepository interface {
	Repository
	// Create creates a new user in the database
	Create(ctx context.Context, user *models.User) error
	// GetByID retrieves a user by their ID
	GetByID(ctx context.Context, id uint) (*models.User, error)
	// GetByEmail retrieves a user by their email address
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	// Update updates an existing user in the database
	Update(ctx context.Context, user *models.User) error
	// Delete removes a user from the database
	Delete(ctx context.Context, id uint) error
	// List retrieves a paginated list of users
	List(ctx context.Context, limit, offset int) ([]*models.User, error)
	// Count returns the total number of users
	Count(ctx context.Context) (int64, error)
	// Exists checks if a user exists by ID
	Exists(ctx context.Context, id uint) (bool, error)
	// ExistsByEmail checks if a user exists by email
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

// userRepository implements the UserRepository interface
type userRepository struct {
	*BaseRepository
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *database.Database) UserRepository {
	return &userRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// getPlaceholder returns the appropriate placeholder for the database driver
func (r *userRepository) getPlaceholder(position int) string {
	if r.BaseRepository.db.Config.Driver == "postgres" {
		return fmt.Sprintf("$%d", position)
	}
	// MySQL and SQLite use ? placeholders
	return "?"
}

// Create creates a new user in the database
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Handle different database drivers
	if r.BaseRepository.db.Config.Driver == "postgres" {
		query := `
			INSERT INTO users (name, email, created_at, updated_at)
			VALUES ($1, $2, $3, $4)
			RETURNING id`

		err := r.GetDB().QueryRowContext(ctx, query, user.Name, user.Email, user.CreatedAt, user.UpdatedAt).Scan(&user.ID)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		// MySQL and SQLite
		query := `
			INSERT INTO users (name, email, created_at, updated_at)
			VALUES (?, ?, ?, ?)`

		result, err := r.GetDB().ExecContext(ctx, query, user.Name, user.Email, user.CreatedAt, user.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get last insert id: %w", err)
		}
		user.ID = uint(id)
	}

	return nil
}

// GetByID retrieves a user by their ID
func (r *userRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	placeholder := r.getPlaceholder(1)
	query := fmt.Sprintf(`
		SELECT id, name, email, created_at, updated_at
		FROM users
		WHERE id = %s`, placeholder)

	user := &models.User{}
	err := r.GetDB().QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by their email address
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	placeholder := r.getPlaceholder(1)
	query := fmt.Sprintf(`
		SELECT id, name, email, created_at, updated_at
		FROM users
		WHERE email = %s`, placeholder)

	user := &models.User{}
	err := r.GetDB().QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with email %s not found", email)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// Update updates an existing user in the database
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	p1, p2, p3, p4 := r.getPlaceholder(1), r.getPlaceholder(2), r.getPlaceholder(3), r.getPlaceholder(4)
	query := fmt.Sprintf(`
		UPDATE users
		SET name = %s, email = %s, updated_at = %s
		WHERE id = %s`, p1, p2, p3, p4)

	user.UpdatedAt = time.Now()

	result, err := r.GetDB().ExecContext(ctx, query, user.Name, user.Email, user.UpdatedAt, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user with id %d not found", user.ID)
	}

	return nil
}

// Delete removes a user from the database
func (r *userRepository) Delete(ctx context.Context, id uint) error {
	placeholder := r.getPlaceholder(1)
	query := fmt.Sprintf(`DELETE FROM users WHERE id = %s`, placeholder)

	result, err := r.GetDB().ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user with id %d not found", id)
	}

	return nil
}

// List retrieves a paginated list of users
func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	p1, p2 := r.getPlaceholder(1), r.getPlaceholder(2)
	query := fmt.Sprintf(`
		SELECT id, name, email, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT %s OFFSET %s`, p1, p2)

	rows, err := r.GetDB().QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over users: %w", err)
	}

	return users, nil
}

// Count returns the total number of users
func (r *userRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM users`

	var count int64
	err := r.GetDB().QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// Exists checks if a user exists by ID
func (r *userRepository) Exists(ctx context.Context, id uint) (bool, error) {
	placeholder := r.getPlaceholder(1)

	var query string
	var exists bool

	if r.BaseRepository.db.Config.Driver == "postgres" {
		query = fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM users WHERE id = %s)`, placeholder)
		err := r.GetDB().QueryRowContext(ctx, query, id).Scan(&exists)
		if err != nil {
			return false, fmt.Errorf("failed to check if user exists: %w", err)
		}
	} else {
		// MySQL and SQLite use COUNT approach
		query = fmt.Sprintf(`SELECT COUNT(*) FROM users WHERE id = %s`, placeholder)
		var count int
		err := r.GetDB().QueryRowContext(ctx, query, id).Scan(&count)
		if err != nil {
			return false, fmt.Errorf("failed to check if user exists: %w", err)
		}
		exists = count > 0
	}

	return exists, nil
}

// ExistsByEmail checks if a user exists by email
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	placeholder := r.getPlaceholder(1)

	var query string
	var exists bool

	if r.BaseRepository.db.Config.Driver == "postgres" {
		query = fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM users WHERE email = %s)`, placeholder)
		err := r.GetDB().QueryRowContext(ctx, query, email).Scan(&exists)
		if err != nil {
			return false, fmt.Errorf("failed to check if user exists by email: %w", err)
		}
	} else {
		// MySQL and SQLite use COUNT approach
		query = fmt.Sprintf(`SELECT COUNT(*) FROM users WHERE email = %s`, placeholder)
		var count int
		err := r.GetDB().QueryRowContext(ctx, query, email).Scan(&count)
		if err != nil {
			return false, fmt.Errorf("failed to check if user exists by email: %w", err)
		}
		exists = count > 0
	}

	return exists, nil
}
