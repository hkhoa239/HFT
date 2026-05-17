package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"quantalpha/internal/database/mocks"
	"quantalpha/internal/models"
	"quantalpha/internal/validator"
)

func TestAlphaHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validator.Init()
	mockDB := mocks.NewMockDB()
	handler := NewAlphaHandler(mockDB)

	userID := uuid.New()
	user := &models.User{ID: userID, Username: "testauthor", Role: models.RoleQR}

	t.Run("Create Alpha", func(t *testing.T) {
		reqBody := models.CreateAlphaRequest{
			Name:        "Test Alpha",
			Description: "Test Description",
			CodeContent: "print('hello')",
		}
		body, _ := json.Marshal(reqBody)

		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("user_id", userID)
			c.Set("user", user)
			c.Next()
		})
		r.POST("/alphas", handler.CreateAlpha)

		req, _ := http.NewRequest("POST", "/alphas", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		
		var resp models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp.Success)
		
		alphaData := resp.Data.(map[string]interface{})
		assert.Equal(t, "Test Alpha", alphaData["name"])
	})

	t.Run("List My Alphas", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("user_id", userID)
			c.Next()
		})
		r.GET("/alphas", handler.ListMyAlphas)

		req, _ := http.NewRequest("GET", "/alphas", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var resp models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp.Success)
		alphas := resp.Data.([]interface{})
		assert.Len(t, alphas, 1)
	})

	t.Run("Update Alpha - Ownership Check", func(t *testing.T) {
		// Create an alpha by another user
		otherUserID := uuid.New()
		alpha, _ := mockDB.CreateAlpha(nil, otherUserID, "Other Alpha", "", "")

		reqBody := models.UpdateAlphaRequest{
			Name: "Malicious Update",
		}
		body, _ := json.Marshal(reqBody)

		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("user_id", userID)
			c.Next()
		})
		r.PUT("/alphas/:id", handler.UpdateAlpha)

		req, _ := http.NewRequest("PUT", "/alphas/"+alpha.ID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
