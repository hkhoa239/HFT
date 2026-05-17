package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"quantalpha/internal/models"
	"quantalpha/internal/utils"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error:   "authorization header required",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error:   "invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := utils.ValidateJWT(tokenString, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error:   "invalid or expired token",
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*utils.JWTClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error:   "invalid token claims",
			})
			c.Abort()
			return
		}

		user := &models.User{
			ID:       claims.UserID,
			Username: claims.Username,
			Role:     models.UserRole(claims.Role),
		}

		c.Set("user", user)
		c.Set("user_id", claims.UserID)
		log.Printf("Authenticated user: %v (ID: %v)", claims.Username, claims.UserID)
		c.Next()
	}
}

func RBACMiddleware(allowedRoles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userVal, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error:   "unauthorized",
			})
			c.Abort()
			return
		}

		user := userVal.(*models.User)

		if user.Role == models.RoleAdmin {
			c.Next()
			return
		}

		for _, role := range allowedRoles {
			if user.Role == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error:   "forbidden: insufficient permissions",
		})
		c.Abort()
	}
}

func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userVal, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error:   "unauthorized",
			})
			c.Abort()
			return
		}

		user := userVal.(*models.User)

		if user.Role != models.RoleAdmin {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Error:   "forbidden: admin only",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
