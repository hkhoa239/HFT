package handlers

import (
	"fmt"
	"math"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"quantalpha/internal/database"
	"quantalpha/internal/models"
	"quantalpha/internal/services"
)

type AnalyticsHandler struct {
	db       database.Querier
	pipeline *services.FactorPipelineService
}

func NewAnalyticsHandler(db database.Querier) *AnalyticsHandler {
	return &AnalyticsHandler{
		db:       db,
		pipeline: services.NewFactorPipelineService(os.Getenv("DATA_DIR")),
	}
}

func (h *AnalyticsHandler) GetCorrelation(c *gin.Context) {
	rawItems, err := h.db.GetCorrelationData(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to fetch correlation data"})
		return
	}

	type alphaData struct {
		name string
		pnl  []float64
	}
	var data []alphaData
	minLength := -1

	for _, item := range rawItems {
		// Convert cumPnL to step returns
		var pnl []float64
		lastVal := 0.0
		for i, p := range item.PnLCurve {
			curr, ok := p["cumPnL"].(float64)
			if !ok {
				// Try parsing from string if it was stored that way
				if s, ok := p["cumPnL"].(string); ok {
					fmt.Sscanf(s, "%f", &curr)
				}
			}
			if i > 0 {
				pnl = append(pnl, curr-lastVal)
			}
			lastVal = curr
		}

		if len(pnl) > 0 {
			data = append(data, alphaData{name: item.AlphaName, pnl: pnl})
			if minLength == -1 || len(pnl) < minLength {
				minLength = len(pnl)
			}
		}
	}

	if len(data) < 2 || minLength <= 1 {
		c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: map[string]interface{}{
			"labels": []string{},
			"matrix": [][]float64{},
		}})
		return
	}

	// 2. Truncate and Calculate Pearson Matrix
	labels := make([]string, len(data))
	for i := range data {
		labels[i] = data[i].name
		data[i].pnl = data[i].pnl[:minLength]
	}

	n := len(data)
	matrix := make([][]float64, n)
	for i := 0; i < n; i++ {
		matrix[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			if i == j {
				matrix[i][j] = 1.0
				continue
			}
			if i > j {
				matrix[i][j] = matrix[j][i]
				continue
			}
			matrix[i][j] = calculatePearson(data[i].pnl, data[j].pnl)
		}
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: map[string]interface{}{
		"labels": labels,
		"matrix": matrix,
	}})
}

func calculatePearson(x, y []float64) float64 {
	n := len(x)
	var sumX, sumY, sumXY, sumX2, sumY2 float64
	for i := 0; i < n; i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
		sumY2 += y[i] * y[i]
	}

	valX := float64(n)*sumX2 - sumX*sumX
	valY := float64(n)*sumY2 - sumY*sumY
	if valX < 0 {
		valX = 0
	}
	if valY < 0 {
		valY = 0
	}

	denom := math.Sqrt(valX * valY)
	if denom == 0 {
		return 0
	}
	return (float64(n)*sumXY - sumX*sumY) / denom
}

func (h *AnalyticsHandler) GetPerformance(c *gin.Context) {
	items, err := h.db.GetPerformance(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: items})
}

func (h *AnalyticsHandler) GetAuditLogs(c *gin.Context) {
	logs, total, err := h.db.ListAuditLogs(c.Request.Context(), 100, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to list audit logs"})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: models.AuditLogResponse{
		Data:  logs,
		Total: total,
	}})
}

// GetDSOverview returns real computed statistics from the database
func (h *AnalyticsHandler) GetDSOverview(c *gin.Context) {
	ctx := c.Request.Context()
	factors, _ := h.db.ListFactors(ctx)
	modelsList, _ := h.db.ListModels(ctx)
	backtests, _ := h.db.ListBacktestRuns(ctx)

	completedCount := 0
	for _, bt := range backtests {
		if bt.Status == "completed" {
			completedCount++
		}
	}

	recordCount, _ := h.pipeline.CountPublishedRows()
	stats := map[string]interface{}{
		"symbol":              "VN30F2112",
		"total_factors":       len(factors),
		"total_models":        len(modelsList),
		"total_backtests":     len(backtests),
		"completed_backtests": completedCount,
		"record_count":        recordCount,
		"tick_size":           0.1,
		"l3_depth":            3,
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: stats})
}

// GetModelMetrics returns training metrics for all models from DB
func (h *AnalyticsHandler) GetModelMetrics(c *gin.Context) {
	ctx := c.Request.Context()
	modelsList, err := h.db.ListModels(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to list models"})
		return
	}

	type modelMetric struct {
		ID              string                 `json:"id"`
		Name            string                 `json:"name"`
		Version         string                 `json:"version"`
		TrainingMetrics map[string]interface{} `json:"training_metrics"`
		CreatedAt       interface{}            `json:"created_at"`
	}

	var results []modelMetric
	for _, m := range modelsList {
		results = append(results, modelMetric{
			ID:              m.ID.String(),
			Name:            m.Name,
			Version:         m.Version,
			TrainingMetrics: m.TrainingMetrics,
			CreatedAt:       m.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: results})
}
