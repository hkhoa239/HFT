package utils

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"quantalpha/internal/models"
)

func TestJWTFlow(t *testing.T) {
	secret := "test-secret-key"
	userID := uuid.New()
	username := "testuser"
	role := models.RoleQR
	expireHour := 1

	// 1. Generate Token
	tokenString, err := GenerateJWT(userID, username, role, secret, expireHour)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// 2. Validate Token
	token, err := ValidateJWT(tokenString, secret)
	assert.NoError(t, err)
	assert.True(t, token.Valid)

	// 3. Verify Claims
	claims, ok := token.Claims.(*JWTClaims)
	assert.True(t, ok)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, string(role), claims.Role)
	assert.Equal(t, "quantalpha", claims.Issuer)
}

func TestJWTInvalidSecret(t *testing.T) {
	secret := "correct-secret"
	wrongSecret := "wrong-secret"
	userID := uuid.New()

	tokenString, _ := GenerateJWT(userID, "user", models.RoleViewer, secret, 1)

	_, err := ValidateJWT(tokenString, wrongSecret)
	assert.Error(t, err)
}

func TestJWTExpired(t *testing.T) {
	secret := "test-secret"
	userID := uuid.New()
	
	// Generate with -1 hour (already expired)
	tokenString, _ := GenerateJWT(userID, "user", models.RoleViewer, secret, -1)

	_, err := ValidateJWT(tokenString, secret)
	assert.Error(t, err)
}
