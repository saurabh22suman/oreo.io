package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// JWTClaims represents the claims stored in JWT tokens
type JWTClaims struct {
	UserID    string `json:"user_id"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// Valid validates the claims
func (c *JWTClaims) Valid() error {
	if c.UserID == "" {
		return errors.New("user ID is required")
	}
	if c.TokenType == "" {
		return errors.New("token type is required")
	}
	return c.RegisteredClaims.Valid()
}

// JWTService interface defines JWT operations
type JWTService interface {
	GenerateTokenPair(userID uuid.UUID) (*TokenPair, error)
	ValidateAccessToken(token string) (*JWTClaims, error)
	RefreshAccessToken(refreshToken string) (*TokenPair, error)
}

// TokenPair represents a pair of access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// jwtServiceImpl implements JWTService
type jwtServiceImpl struct {
	secretKey            []byte
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string) JWTService {
	accessDuration := 15 * time.Minute    // Default 15 minutes
	refreshDuration := 7 * 24 * time.Hour // Default 7 days

	// Parse durations from environment if available
	if accessEnv := os.Getenv("JWT_ACCESS_EXPIRY"); accessEnv != "" {
		if d, err := time.ParseDuration(accessEnv); err == nil {
			accessDuration = d
		}
	}

	if refreshEnv := os.Getenv("JWT_REFRESH_EXPIRY"); refreshEnv != "" {
		if d, err := time.ParseDuration(refreshEnv); err == nil {
			refreshDuration = d
		}
	}

	return &jwtServiceImpl{
		secretKey:            []byte(secretKey),
		accessTokenDuration:  accessDuration,
		refreshTokenDuration: refreshDuration,
	}
}

// GenerateTokenPair generates both access and refresh tokens
func (j *jwtServiceImpl) GenerateTokenPair(userID uuid.UUID) (*TokenPair, error) {
	// Generate access token
	accessClaims := &JWTClaims{
		UserID:    userID.String(),
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "oreo.io",
			Subject:   userID.String(),
			ID:        uuid.New().String(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(j.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshClaims := &JWTClaims{
		UserID:    userID.String(),
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.refreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "oreo.io",
			Subject:   userID.String(),
			ID:        uuid.New().String(),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(j.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}

// ValidateAccessToken validates an access token and returns the claims
func (j *jwtServiceImpl) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	return j.validateToken(tokenString, "access")
}

// RefreshAccessToken generates a new access token using a refresh token
func (j *jwtServiceImpl) RefreshAccessToken(refreshToken string) (*TokenPair, error) {
	claims, err := j.validateToken(refreshToken, "refresh")
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in token: %w", err)
	}

	return j.GenerateTokenPair(userID)
}

// validateToken is a helper method to validate tokens
func (j *jwtServiceImpl) validateToken(tokenString, expectedType string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("expected %s token, got %s", expectedType, claims.TokenType)
	}

	return claims, nil
}
