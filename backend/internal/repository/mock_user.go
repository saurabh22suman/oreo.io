package repository

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/saurabh22suman/oreo.io/internal/models"
)

// mockUserRepository implements UserRepository interface for development
type mockUserRepository struct {
	users map[string]*models.User  // email -> user
	mu    sync.RWMutex
}

// NewMockUserRepository creates a new mock user repository
func NewMockUserRepository() UserRepository {
	return &mockUserRepository{
		users: make(map[string]*models.User),
	}
}

// Create creates a new user
func (r *mockUserRepository) Create(ctx context.Context, user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.Email]; exists {
		return ErrUserAlreadyExists
	}

	// Generate UUID if not set
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	// Store user
	r.users[user.Email] = user
	return nil
}

// GetByID retrieves a user by ID
func (r *mockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.ID == id {
			return user, nil
		}
	}

	return nil, ErrUserNotFound
}

// GetByEmail retrieves a user by email
func (r *mockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[email]
	if !exists {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// GetByGoogleID retrieves a user by Google ID
func (r *mockUserRepository) GetByGoogleID(ctx context.Context, googleID string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.GoogleID == googleID {
			return user, nil
		}
	}

	return nil, ErrUserNotFound
}

// Update updates a user
func (r *mockUserRepository) Update(ctx context.Context, user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.Email]; !exists {
		return ErrUserNotFound
	}

	r.users[user.Email] = user
	return nil
}

// Delete deletes a user by ID
func (r *mockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for email, user := range r.users {
		if user.ID == id {
			delete(r.users, email)
			return nil
		}
	}

	return ErrUserNotFound
}

// List retrieves users with pagination
func (r *mockUserRepository) List(ctx context.Context, offset, limit int) ([]*models.User, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	total := len(r.users)
	users := make([]*models.User, 0, len(r.users))

	i := 0
	for _, user := range r.users {
		if i >= offset && len(users) < limit {
			users = append(users, user)
		}
		i++
	}

	return users, total, nil
}

// EmailExists checks if an email already exists
func (r *mockUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.users[email]
	return exists, nil
}
