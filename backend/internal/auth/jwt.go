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

// JWTService handles JWT token operations
type JWTService struct {
	secretKey            []byte
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string) *JWTService {
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

	return &JWTService{
		secretKey:            []byte(secretKey),
		accessTokenDuration:  accessDuration,
		refreshTokenDuration: refreshDuration,
	}
}

// GenerateTokenPair generates both access and refresh tokens
func (s *JWTService) GenerateTokenPair(userID uuid.UUID) (accessToken, refreshToken string, err error) {
	now := time.Now()

	// Generate access token
	accessClaims := &JWTClaims{
		UserID:    userID.String(),
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "oreo.io",
		},
	}

	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessTokenObj.SignedString(s.secretKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshClaims := &JWTClaims{
		UserID:    userID.String(),
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "oreo.io",
		},
	}

	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTokenObj.SignedString(s.secretKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// ValidateAccessToken validates an access token and returns claims
func (s *JWTService) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	return s.validateToken(tokenString, "access")
}

// ValidateRefreshToken validates a refresh token and returns claims
func (s *JWTService) ValidateRefreshToken(tokenString string) (*JWTClaims, error) {
	return s.validateToken(tokenString, "refresh")
}

// validateToken validates a token with specific type
func (s *JWTService) validateToken(tokenString, expectedType string) (*JWTClaims, error) {
	if tokenString == "" {
		return nil, errors.New("token is required")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Validate token type
	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("invalid token type: expected %s, got %s", expectedType, claims.TokenType)
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token using a valid refresh token
func (s *JWTService) RefreshAccessToken(refreshToken string) (string, error) {
	claims, err := s.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return "", fmt.Errorf("invalid user ID in token: %w", err)
	}

	// Generate new access token only (refresh token remains the same)
	now := time.Now()
	accessClaims := &JWTClaims{
		UserID:    userID.String(),
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "oreo.io",
		},
	}

	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err := accessTokenObj.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to generate new access token: %w", err)
	}

	return accessToken, nil
}

// ExtractUserID extracts user ID from a valid access token
func (s *JWTService) ExtractUserID(tokenString string) (uuid.UUID, error) {
	claims, err := s.ValidateAccessToken(tokenString)
	if err != nil {
		return uuid.Nil, err
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID in token: %w", err)
	}

	return userID, nil
}
