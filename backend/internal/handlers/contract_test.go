package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"quantalpha/internal/database/mocks"
	"quantalpha/internal/models"
)

func TestAPIContractStability(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := mocks.NewMockDB()
	alphaHandler := NewAlphaHandler(mockDB)

	userID := uuid.New()
	alphaID := uuid.New()
	user := &models.User{ID: userID, Username: "contract_user", Role: models.RoleQR}
	
	alpha := &models.Alpha{
		ID:          alphaID,
		AuthorID:    userID,
		Name:        "Contract Alpha",
		Description: "Stable Schema",
		CodeContent: "pass",
		Status:      models.AlphaStatusDraft,
	}
	mockDB.Users[userID] = user
	mockDB.Alphas[alphaID] = alpha

	t.Run("Alpha Response Structure", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("user_id", userID)
			c.Set("user", user)
			c.Next()
		})
		r.GET("/alphas/:id", alphaHandler.GetAlpha)

		req, _ := http.NewRequest("GET", "/alphas/"+alphaID.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var raw map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &raw)

		// 1. Root structure
		assert.Contains(t, raw, "success")
		assert.Contains(t, raw, "data")

		data := raw["data"].(map[string]interface{})

		// 2. REQUIRED fields for Frontend
		// If these disappear or rename, Frontend breaks.
		requiredFields := []string{"id", "author_id", "name", "status", "created_at"}
		for _, f := range requiredFields {
			assert.Contains(t, data, f, "Missing required field: "+f)
		}

		// 3. ENUM validation
		assert.Equal(t, "draft", data["status"], "Enum value mismatch")

		// 4. NULLABLE handling
		// description can be NULL, but it should be present as null or empty string
		assert.Contains(t, data, "description")
	})
}
