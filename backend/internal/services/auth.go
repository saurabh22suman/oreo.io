package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/saurabh22suman/oreo.io/internal/auth"
	"github.com/saurabh22suman/oreo.io/internal/models"
	"github.com/saurabh22suman/oreo.io/internal/repository"
)

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	User   models.PublicUser `json:"user"`
	Tokens TokenPair         `json:"tokens"`
}

// AuthService defines the interface for authentication operations
type AuthService interface {
	Register(ctx context.Context, req *models.CreateUserRequest) (*AuthResponse, error)
	Login(ctx context.Context, req *models.LoginRequest) (*AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
	GetUserFromToken(ctx context.Context, token string) (*models.User, error)
	Logout(ctx context.Context, userID uuid.UUID) error
}

// authService implements AuthService interface
type authService struct {
	userRepo   repository.UserRepository
	jwtService auth.JWTService
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo repository.UserRepository, jwtService auth.JWTService) AuthService {
	return &authService{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

// Register creates a new user account and returns auth tokens
func (s *authService) Register(ctx context.Context, req *models.CreateUserRequest) (*AuthResponse, error) {
	// Create user model from request
	user := &models.User{
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password,
	}

	// Validate user data
	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create user in repository
	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return nil, fmt.Errorf("user with email %s already exists", req.Email)
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	tokenPair, err := s.jwtService.GenerateTokenPair(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &AuthResponse{
		User: user.PublicUser(),
		Tokens: TokenPair{
			AccessToken:  tokenPair.AccessToken,
			RefreshToken: tokenPair.RefreshToken,
		},
	}, nil
}

// Login authenticates a user and returns auth tokens
func (s *authService) Login(ctx context.Context, req *models.LoginRequest) (*AuthResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check password
	if !user.CheckPassword(req.Password) {
		return nil, errors.New("invalid email or password")
	}

	// Generate tokens
	tokenPair, err := s.jwtService.GenerateTokenPair(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &AuthResponse{
		User: user.PublicUser(),
		Tokens: TokenPair{
			AccessToken:  tokenPair.AccessToken,
			RefreshToken: tokenPair.RefreshToken,
		},
	}, nil
}

// RefreshToken generates a new access token from a refresh token
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	// Validate refresh token and get new access token
	tokenPair, err := s.jwtService.RefreshAccessToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}

	return tokenPair.AccessToken, nil
}

// GetUserFromToken retrieves a user from an access token
func (s *authService) GetUserFromToken(ctx context.Context, token string) (*models.User, error) {
	// Validate access token
	claims, err := s.jwtService.ValidateAccessToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Parse user ID
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in token: %w", err)
	}

	// Get user from repository
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// Logout handles user logout (placeholder for future token blacklisting)
func (s *authService) Logout(ctx context.Context, userID uuid.UUID) error {
	// TODO: Implement token blacklisting with Redis
	// For now, logout is handled client-side by removing tokens
	return nil
}
