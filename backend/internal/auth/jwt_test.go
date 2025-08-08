package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTService_GenerateTokenPair(t *testing.T) {
	service := NewJWTService("test-secret-key-at-least-32-characters-long")
	userID := uuid.New()

	accessToken, refreshToken, err := service.GenerateTokenPair(userID)

	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.NotEqual(t, accessToken, refreshToken)
}

func TestJWTService_ValidateAccessToken(t *testing.T) {
	service := NewJWTService("test-secret-key-at-least-32-characters-long")
	userID := uuid.New()

	// Generate token
	accessToken, _, err := service.GenerateTokenPair(userID)
	require.NoError(t, err)

	// Validate token
	claims, err := service.ValidateAccessToken(accessToken)
	require.NoError(t, err)

	assert.Equal(t, userID.String(), claims.UserID)
	assert.Equal(t, "access", claims.TokenType)
	assert.True(t, time.Now().Before(claims.ExpiresAt.Time))
}

func TestJWTService_ValidateRefreshToken(t *testing.T) {
	service := NewJWTService("test-secret-key-at-least-32-characters-long")
	userID := uuid.New()

	// Generate token
	_, refreshToken, err := service.GenerateTokenPair(userID)
	require.NoError(t, err)

	// Validate token
	claims, err := service.ValidateRefreshToken(refreshToken)
	require.NoError(t, err)

	assert.Equal(t, userID.String(), claims.UserID)
	assert.Equal(t, "refresh", claims.TokenType)
	assert.True(t, time.Now().Before(claims.ExpiresAt.Time))
}

func TestJWTService_InvalidToken(t *testing.T) {
	service := NewJWTService("test-secret-key-at-least-32-characters-long")

	// Test invalid token
	_, err := service.ValidateAccessToken("invalid.token.here")
	assert.Error(t, err)

	// Test empty token
	_, err = service.ValidateAccessToken("")
	assert.Error(t, err)
}

func TestJWTService_ExpiredToken(t *testing.T) {
	// Create service with very short expiry for testing
	service := &JWTService{
		secretKey:            []byte("test-secret-key-at-least-32-characters-long"),
		accessTokenDuration:  time.Millisecond, // Very short expiry
		refreshTokenDuration: time.Millisecond,
	}

	userID := uuid.New()

	// Generate token
	accessToken, _, err := service.GenerateTokenPair(userID)
	require.NoError(t, err)

	// Wait for expiry
	time.Sleep(time.Millisecond * 10)

	// Validate expired token
	_, err = service.ValidateAccessToken(accessToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestJWTService_WrongTokenType(t *testing.T) {
	service := NewJWTService("test-secret-key-at-least-32-characters-long")
	userID := uuid.New()

	// Generate tokens
	accessToken, refreshToken, err := service.GenerateTokenPair(userID)
	require.NoError(t, err)

	// Try to validate access token as refresh token
	_, err = service.ValidateRefreshToken(accessToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token type")

	// Try to validate refresh token as access token
	_, err = service.ValidateAccessToken(refreshToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token type")
}

func TestJWTService_RefreshAccessToken(t *testing.T) {
	service := NewJWTService("test-secret-key-at-least-32-characters-long")
	userID := uuid.New()

	// Generate initial tokens
	_, refreshToken, err := service.GenerateTokenPair(userID)
	require.NoError(t, err)

	// Refresh access token
	newAccessToken, err := service.RefreshAccessToken(refreshToken)
	require.NoError(t, err)
	assert.NotEmpty(t, newAccessToken)

	// Validate new access token
	claims, err := service.ValidateAccessToken(newAccessToken)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), claims.UserID)
}

func TestJWTClaims_Valid(t *testing.T) {
	// Valid claims
	claims := &JWTClaims{
		UserID:    uuid.New().String(),
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	err := claims.Valid()
	assert.NoError(t, err)

	// Invalid - missing user ID
	claims.UserID = ""
	err = claims.Valid()
	assert.Error(t, err)

	// Invalid - missing token type
	claims.UserID = uuid.New().String()
	claims.TokenType = ""
	err = claims.Valid()
	assert.Error(t, err)
}
