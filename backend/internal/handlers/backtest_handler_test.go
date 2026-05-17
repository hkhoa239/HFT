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
	dbmocks "quantalpha/internal/database/mocks"
	"quantalpha/internal/models"
	redismocks "quantalpha/internal/redis/mocks"
	"quantalpha/internal/validator"
)

func TestBacktestHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validator.Init()
	mockDB := dbmocks.NewMockDB()
	mockProd := redismocks.NewMockProducer()
	handler := NewBacktestHandler(mockDB, mockProd)

	userID := uuid.New()
	user := &models.User{ID: userID, Username: "quant_researcher", Role: models.RoleQR}
	mockDB.Users[userID] = user

	alphaID := uuid.New()
	mockDB.Alphas[alphaID] = &models.Alpha{
		ID:          alphaID,
		AuthorID:    userID,
		Name:        "Test Alpha",
		CodeContent: "def alpha(): pass",
	}

	t.Run("Run Backtest - Success", func(t *testing.T) {
		reqBody := models.RunBacktestRequest{
			AlphaID: alphaID.String(),
			Params: &models.BacktestParams{
				Start:   "2023-01-01",
				End:     "2023-01-31",
				Capital: 100000.0,
			},
		}
		body, _ := json.Marshal(reqBody)

		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("user_id", userID)
			c.Set("user", user)
			c.Next()
		})
		r.POST("/backtest", handler.RunBacktest)

		req, _ := http.NewRequest("POST", "/backtest", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		
		var resp models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp.Success)
		
		// Verify job published to redis
		assert.Len(t, mockProd.PublishedJobs, 1)
		assert.Equal(t, alphaID.String(), mockProd.PublishedJobs[0].AlphaID)
		
		// Verify audit log created
		assert.Len(t, mockDB.AuditLogs, 1)
		assert.Equal(t, models.AuditRun, mockDB.AuditLogs[0].Action)
	})

	t.Run("Get Backtest Status", func(t *testing.T) {
		backtestID := uuid.New()
		mockDB.BacktestRuns[backtestID] = &models.BacktestRun{
			ID:      backtestID,
			AlphaID: alphaID,
			Status:  models.JobStatusRunning,
		}

		r := gin.New()
		r.GET("/backtest/:job_id", handler.GetBacktestStatus)

		req, _ := http.NewRequest("GET", "/backtest/"+backtestID.String(), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp.Success)
		
		data := resp.Data.(map[string]interface{})
		assert.Equal(t, string(models.JobStatusRunning), data["status"])
	})
}
