package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"quantalpha/internal/models"
	"quantalpha/internal/utils"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret"
	userID := uuid.New()
	username := "testuser"
	role := models.RoleQR

	t.Run("Valid Token", func(t *testing.T) {
		token, _ := utils.GenerateJWT(userID, username, role, secret, 1)

		r := gin.New()
		r.Use(AuthMiddleware(secret))
		r.GET("/test", func(c *gin.Context) {
			user, _ := c.Get("user")
			u := user.(*models.User)
			assert.Equal(t, userID, u.ID)
			assert.Equal(t, username, u.Username)
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Missing Header", func(t *testing.T) {
		r := gin.New()
		r.Use(AuthMiddleware(secret))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		r := gin.New()
		r.Use(AuthMiddleware(secret))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestRBACMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Allowed Role", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("user", &models.User{Role: models.RoleQR})
			c.Next()
		})
		r.Use(RBACMiddleware(models.RoleQR))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Forbidden Role", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("user", &models.User{Role: models.RoleViewer})
			c.Next()
		})
		r.Use(RBACMiddleware(models.RoleQR))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Admin Bypass", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("user", &models.User{Role: models.RoleAdmin})
			c.Next()
		})
		r.Use(RBACMiddleware(models.RoleQR))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
