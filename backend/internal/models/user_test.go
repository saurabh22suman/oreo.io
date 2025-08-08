package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid user",
			user: User{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "invalid email - empty",
			user: User{
				Email:    "",
				Name:     "Test User",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "email is required",
		},
		{
			name: "invalid email - format",
			user: User{
				Email:    "invalid-email",
				Name:     "Test User",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name: "invalid name - empty",
			user: User{
				Email:    "test@example.com",
				Name:     "",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "invalid password - too short",
			user: User{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "123",
			},
			wantErr: true,
			errMsg:  "password must be at least 6 characters",
		},
		{
			name: "invalid name - too long",
			user: User{
				Email:    "test@example.com",
				Name:     "This is a very long name that exceeds the maximum allowed length for a user name field in our application",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "name must be less than 100 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUser_HashPassword(t *testing.T) {
	user := &User{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "plaintext123",
	}

	err := user.HashPassword()

	assert.NoError(t, err)
	assert.NotEqual(t, "plaintext123", user.Password)
	assert.True(t, len(user.Password) > 50)   // bcrypt hashes are long
	assert.Contains(t, user.Password, "$2a$") // bcrypt prefix
}

func TestUser_CheckPassword(t *testing.T) {
	user := &User{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "plaintext123",
	}

	// Hash the password first
	err := user.HashPassword()
	require.NoError(t, err)

	// Test correct password
	assert.True(t, user.CheckPassword("plaintext123"))

	// Test incorrect password
	assert.False(t, user.CheckPassword("wrongpassword"))
	assert.False(t, user.CheckPassword(""))
}

func TestUser_BeforeCreate(t *testing.T) {
	user := &User{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}

	err := user.BeforeCreate()

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.NotEqual(t, "password123", user.Password) // Should be hashed
	assert.False(t, user.CreatedAt.IsZero())
	assert.False(t, user.UpdatedAt.IsZero())
}

func TestUser_IsValidRole(t *testing.T) {
	tests := []struct {
		role  string
		valid bool
	}{
		{"admin", true},
		{"editor", true},
		{"reviewer", true},
		{"viewer", true},
		{"invalid", false},
		{"", false},
		{"ADMIN", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			assert.Equal(t, tt.valid, IsValidRole(tt.role))
		})
	}
}

func TestUser_TableName(t *testing.T) {
	user := &User{}
	assert.Equal(t, "users", user.TableName())
}

func TestUser_PublicUser(t *testing.T) {
	user := &User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Name:      "Test User",
		Password:  "hashed_password",
		GoogleID:  "google123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	publicUser := user.PublicUser()

	assert.Equal(t, user.ID, publicUser.ID)
	assert.Equal(t, user.Email, publicUser.Email)
	assert.Equal(t, user.Name, publicUser.Name)
	assert.Equal(t, user.GoogleID, publicUser.GoogleID)
	assert.Equal(t, user.CreatedAt, publicUser.CreatedAt)
	assert.Equal(t, user.UpdatedAt, publicUser.UpdatedAt)
	// Password should not be accessible in PublicUser struct
}
