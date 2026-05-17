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

func TestAnalyticsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB := mocks.NewMockDB()
	handler := NewAnalyticsHandler(mockDB)

	userID := uuid.New()
	author := &models.User{ID: userID, Username: "quant_pro", Role: models.RoleQR}
	mockDB.Users[userID] = author

	alphaID := uuid.New()
	mockDB.Alphas[alphaID] = &models.Alpha{
		ID:       alphaID,
		AuthorID: userID,
		Name:     "Momentum Alpha",
	}

	t.Run("Get Performance - No Data", func(t *testing.T) {
		r := gin.New()
		r.GET("/performance", handler.GetPerformance)

		req, _ := http.NewRequest("GET", "/performance", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp.Success)
		assert.Len(t, resp.Data.([]interface{}), 0)
	})

	t.Run("Get Performance - With Data", func(t *testing.T) {
		// Add a completed backtest run
		backtestID := uuid.New()
		mockDB.BacktestRuns[backtestID] = &models.BacktestRun{
			ID:      backtestID,
			AlphaID: alphaID,
			Status:  models.JobStatusCompleted,
			Metrics: map[string]interface{}{
				"total_pnl":    15.5,
				"sharpe_ratio": 2.1,
			},
		}

		r := gin.New()
		r.GET("/performance", handler.GetPerformance)

		req, _ := http.NewRequest("GET", "/performance", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp.Success)
		
		items := resp.Data.([]interface{})
		assert.Len(t, items, 1)
		
		item := items[0].(map[string]interface{})
		assert.Equal(t, "Momentum Alpha", item["alpha_name"])
		assert.Equal(t, 15.5, item["total_return"])
		assert.Equal(t, 2.1, item["sharpe"])
	})
}
