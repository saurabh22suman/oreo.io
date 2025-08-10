package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("rate limit middleware exists", func(t *testing.T) {
		middleware := RateLimit()
		assert.NotNil(t, middleware)
	})

	t.Run("allows requests within limit", func(t *testing.T) {
		router := gin.New()
		router.Use(RateLimit())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("require auth middleware exists", func(t *testing.T) {
		middleware := RequireAuth()
		assert.NotNil(t, middleware)
	})

	t.Run("blocks requests without auth", func(t *testing.T) {
		router := gin.New()
		router.Use(RequireAuth())
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestRequireAuthWithService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("middleware creation with nil service", func(t *testing.T) {
		middleware := RequireAuthWithService(nil)
		assert.NotNil(t, middleware)
	})

	t.Run("blocks requests without authorization header", func(t *testing.T) {
		router := gin.New()
		router.Use(RequireAuthWithService(nil))
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("blocks requests with malformed authorization header", func(t *testing.T) {
		router := gin.New()
		router.Use(RequireAuthWithService(nil))
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
