package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"quantalpha/internal/database/mocks"
	"quantalpha/internal/models"
	redismocks "quantalpha/internal/redis/mocks"
	"quantalpha/internal/validator"
)

func TestConcurrency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validator.Init()

	mockDB := mocks.NewMockDB()
	mockProd := redismocks.NewMockProducer()
	alphaHandler := NewAlphaHandler(mockDB)
	backtestHandler := NewBacktestHandler(mockDB, mockProd)

	userID := uuid.New()
	user := &models.User{ID: userID, Username: "concurrent_user", Role: models.RoleQR}
	mockDB.Users[userID] = user

	t.Run("Parallel Alpha Submissions", func(t *testing.T) {
		const numRequests = 50
		var wg sync.WaitGroup
		wg.Add(numRequests)

		for i := 0; i < numRequests; i++ {
			go func(idx int) {
				defer wg.Done()
				
				reqBody := models.CreateAlphaRequest{
					Name:        "Concurrent Alpha",
					Description: "Parallel testing",
					CodeContent: "pass",
				}
				body, _ := json.Marshal(reqBody)

				r := gin.New()
				r.Use(func(c *gin.Context) {
					c.Set("user_id", userID)
					c.Set("user", user)
					c.Next()
				})
				r.POST("/alphas", alphaHandler.CreateAlpha)

				req, _ := http.NewRequest("POST", "/alphas", bytes.NewBuffer(body))
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)

				assert.Equal(t, http.StatusCreated, w.Code)
			}(i)
		}

		wg.Wait()
		
		alphas, _ := mockDB.ListAlphasByAuthor(nil, userID)
		assert.Equal(t, numRequests, len(alphas), "All parallel alpha creations should persist")
	})

	t.Run("Parallel Backtest Runs - With Idempotency", func(t *testing.T) {
		alphaID := uuid.New()
		mockDB.Alphas[alphaID] = &models.Alpha{ID: alphaID, AuthorID: userID, Name: "Shared Alpha", CodeContent: "pass"}

		const numRequests = 20
		var wg sync.WaitGroup
		wg.Add(numRequests)

		for i := 0; i < numRequests; i++ {
			go func() {
				defer wg.Done()
				
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
				r.POST("/backtest/run", backtestHandler.RunBacktest)

				req, _ := http.NewRequest("POST", "/backtest/run", bytes.NewBuffer(body))
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)

				// Some will be 201 Created, others will be 409 Conflict due to idempotency
				assert.Contains(t, []int{http.StatusCreated, http.StatusConflict}, w.Code)
			}()
		}

		wg.Wait()
		
		// Verify only ONE job was actually created in DB for this alpha/params combination
		runs, _ := mockDB.ListBacktestRuns(nil)
		count := 0
		for _, r := range runs {
			if r.AlphaID == alphaID {
				count++
			}
		}
		assert.Equal(t, 1, count, "Only one backtest run should be created despite parallel requests")
	})
}
