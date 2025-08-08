package models

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	Password  string    `json:"-" db:"password_hash"` // Never include in JSON
	GoogleID  string    `json:"google_id,omitempty" db:"google_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// PublicUser represents a user without sensitive information
type PublicUser struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	GoogleID  string    `json:"google_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserRole represents valid user roles
const (
	RoleAdmin    = "admin"
	RoleEditor   = "editor"
	RoleReviewer = "reviewer"
	RoleViewer   = "viewer"
)

// Valid roles list
var validRoles = []string{RoleAdmin, RoleEditor, RoleReviewer, RoleViewer}

// Validate validates the user model
func (u *User) Validate() error {
	// Check email
	if strings.TrimSpace(u.Email) == "" {
		return errors.New("email is required")
	}

	// Validate email format
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	if matched, _ := regexp.MatchString(emailRegex, u.Email); !matched {
		return errors.New("invalid email format")
	}

	// Check name
	if strings.TrimSpace(u.Name) == "" {
		return errors.New("name is required")
	}

	if len(u.Name) > 100 {
		return errors.New("name must be less than 100 characters")
	}

	// Check password (only if not empty - for updates)
	if u.Password != "" && len(u.Password) < 6 {
		return errors.New("password must be at least 6 characters")
	}

	return nil
}

// HashPassword hashes the user's password using bcrypt
func (u *User) HashPassword() error {
	if u.Password == "" {
		return errors.New("password cannot be empty")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(hashedBytes)
	return nil
}

// CheckPassword checks if the provided password matches the hashed password
func (u *User) CheckPassword(password string) bool {
	if password == "" || u.Password == "" {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// BeforeCreate prepares the user model before database insertion
func (u *User) BeforeCreate() error {
	// Generate UUID if not set
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}

	// Set timestamps
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now

	// Hash password if not already hashed
	if u.Password != "" && !strings.HasPrefix(u.Password, "$2a$") {
		if err := u.HashPassword(); err != nil {
			return err
		}
	}

	return nil
}

// BeforeUpdate prepares the user model before database update
func (u *User) BeforeUpdate() error {
	u.UpdatedAt = time.Now()
	return nil
}

// PublicUser returns a user struct without sensitive information
func (u *User) PublicUser() PublicUser {
	return PublicUser{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		GoogleID:  u.GoogleID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// TableName returns the table name for the User model
func (u *User) TableName() string {
	return "users"
}

// IsValidRole checks if a role is valid
func IsValidRole(role string) bool {
	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	return false
}

// CreateUserRequest represents the request structure for user creation
type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required,min=1,max=100"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest represents the request structure for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UpdateUserRequest represents the request structure for user updates
type UpdateUserRequest struct {
	Name     *string `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Password *string `json:"password,omitempty" binding:"omitempty,min=6"`
}
