package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/saurabh22suman/oreo.io/internal/models"
)

// ErrUserNotFound is returned when a user is not found
var ErrUserNotFound = errors.New("user not found")

// ErrUserAlreadyExists is returned when trying to create a user that already exists
var ErrUserAlreadyExists = errors.New("user already exists")

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByGoogleID(ctx context.Context, googleID string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*models.User, int, error)
	EmailExists(ctx context.Context, email string) (bool, error)
}

// userRepository implements UserRepository interface
type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user in the database
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	// Prepare user for creation (generate ID, hash password, set timestamps)
	if err := user.BeforeCreate(); err != nil {
		return fmt.Errorf("failed to prepare user for creation: %w", err)
	}

	// Validate user
	if err := user.Validate(); err != nil {
		return fmt.Errorf("user validation failed: %w", err)
	}

	// Check if email already exists
	exists, err := r.EmailExists(ctx, user.Email)
	if err != nil {
		return fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return ErrUserAlreadyExists
	}

	// Insert user into database
	query := `
		INSERT INTO users (id, email, name, password_hash, google_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.Password,
		user.GoogleID,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, name, password_hash, google_id, created_at, updated_at
		FROM users 
		WHERE id = $1`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Password,
		&user.GoogleID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, name, password_hash, google_id, created_at, updated_at
		FROM users 
		WHERE email = $1`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Password,
		&user.GoogleID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// GetByGoogleID retrieves a user by Google ID
func (r *userRepository) GetByGoogleID(ctx context.Context, googleID string) (*models.User, error) {
	query := `
		SELECT id, email, name, password_hash, google_id, created_at, updated_at
		FROM users 
		WHERE google_id = $1`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, googleID).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Password,
		&user.GoogleID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by Google ID: %w", err)
	}

	return user, nil
}

// Update updates an existing user
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	// Prepare user for update (set updated timestamp)
	if err := user.BeforeUpdate(); err != nil {
		return fmt.Errorf("failed to prepare user for update: %w", err)
	}

	// Validate user
	if err := user.Validate(); err != nil {
		return fmt.Errorf("user validation failed: %w", err)
	}

	query := `
		UPDATE users 
		SET email = $2, name = $3, password_hash = $4, google_id = $5, updated_at = $6
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.Password,
		user.GoogleID,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// Delete deletes a user by ID
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// List retrieves users with pagination
func (r *userRepository) List(ctx context.Context, offset, limit int) ([]*models.User, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM users`
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user count: %w", err)
	}

	// Get users with pagination
	query := `
		SELECT id, email, name, password_hash, google_id, created_at, updated_at
		FROM users 
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.Password,
			&user.GoogleID,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating users: %w", err)
	}

	return users, total, nil
}

// EmailExists checks if an email already exists in the database
func (r *userRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return exists, nil
}
