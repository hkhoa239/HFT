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
	redismocks "quantalpha/internal/redis/mocks"
	"quantalpha/internal/validator"
)

func TestBacktestConsistency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validator.Init()

	userID := uuid.New()
	alphaID := uuid.New()
	user := &models.User{ID: userID, Username: "testuser", Role: models.RoleQR}

	t.Run("Redis Failure - Should Cleanup DB Record", func(t *testing.T) {
		mockDB := mocks.NewMockDB()
		mockProd := redismocks.NewMockProducer()
		mockProd.ShouldFail = true // Simulate Redis failure
		
		handler := NewBacktestHandler(mockDB, mockProd)
		mockDB.Users[userID] = user
		mockDB.Alphas[alphaID] = &models.Alpha{ID: alphaID, AuthorID: userID, Name: "Test Alpha", CodeContent: "pass"}

		reqBody := models.RunBacktestRequest{
			AlphaID: alphaID.String(),
			Params: &models.BacktestParams{Start: "2023-01-01", End: "2023-01-02", Capital: 1000},
		}
		body, _ := json.Marshal(reqBody)

		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("user_id", userID)
			c.Set("user", user)
			c.Next()
		})
		r.POST("/backtest/run", handler.RunBacktest)

		req, _ := http.NewRequest("POST", "/backtest/run", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Verification:
		// 1. Status should be 500 (Internal Server Error)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		
		// 2. IMPORTANT: DB record should NOT exist or should be marked as failed
		// In current implementation, it probably still exists as 'pending'.
		// We want to verify it was handled.
		runs, _ := mockDB.ListBacktestRuns(nil)
		
		// If we haven't implemented cleanup yet, this will fail if we expect 0.
		// Let's see what happens.
		assert.Len(t, runs, 0, "DB record should be cleaned up if Redis publish fails")
	})

	t.Run("Duplicate Submission - Should Fail with Conflict", func(t *testing.T) {
		mockDB := mocks.NewMockDB()
		mockProd := redismocks.NewMockProducer()
		handler := NewBacktestHandler(mockDB, mockProd)
		
		mockDB.Users[userID] = user
		mockDB.Alphas[alphaID] = &models.Alpha{ID: alphaID, AuthorID: userID, Name: "Test Alpha", CodeContent: "pass"}

		reqBody := models.RunBacktestRequest{
			AlphaID: alphaID.String(),
			Params: &models.BacktestParams{Start: "2023-01-01", End: "2023-01-02", Capital: 1000},
		}
		body, _ := json.Marshal(reqBody)

		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("user_id", userID)
			c.Set("user", user)
			c.Next()
		})
		r.POST("/backtest/run", handler.RunBacktest)

		// 1. First submission - Success
		req1, _ := http.NewRequest("POST", "/backtest/run", bytes.NewBuffer(body))
		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusCreated, w1.Code)

		// 2. Second submission - Conflict
		req2, _ := http.NewRequest("POST", "/backtest/run", bytes.NewBuffer(body))
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusConflict, w2.Code)
		
		var resp models.APIResponse
		json.Unmarshal(w2.Body.Bytes(), &resp)
		assert.Equal(t, "A backtest with these parameters is already in progress", resp.Error)
	})
}
