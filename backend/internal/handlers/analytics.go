package handlers

import (
	"fmt"
	"log"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"quantalpha/internal/database"
	"quantalpha/internal/models"
)

type AnalyticsHandler struct {
	db *database.Queries
}

func NewAnalyticsHandler(db *database.Queries) *AnalyticsHandler {
	return &AnalyticsHandler{db: db}
}

func (h *AnalyticsHandler) GetCorrelation(c *gin.Context) {
	// 1. Fetch latest 10 completed backtests with pnl curves
	query := `
		WITH LatestBacktests AS (
			SELECT DISTINCT ON (alpha_id) 
				alpha_id, metrics, created_at
			FROM backtest_runs
			WHERE status = 'completed' AND metrics->'pnl_curve' IS NOT NULL
			ORDER BY alpha_id, created_at DESC
			LIMIT 10
		)
		SELECT 
			a.name as alpha_name,
			lb.metrics->'pnl_curve' as pnl_curve
		FROM LatestBacktests lb
		JOIN alphas a ON lb.alpha_id = a.id
		ORDER BY lb.created_at DESC
	`

	rows, err := h.db.GetDB().Query(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to fetch correlation data"})
		return
	}
	defer rows.Close()

	type alphaData struct {
		name   string
		pnl    []float64
	}
	var data []alphaData
	minLength := -1

	for rows.Next() {
		var name string
		var rawCurve []map[string]interface{}
		if err := rows.Scan(&name, &rawCurve); err != nil {
			continue
		}

		// Convert cumPnL to step returns
		var pnl []float64
		lastVal := 0.0
		for i, p := range rawCurve {
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
			data = append(data, alphaData{name: name, pnl: pnl})
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

	// But actually the PM Dashboard expects a matrix. 
	// I'll return both or update CorrelationResponse to include Matrix.
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
	if valX < 0 { valX = 0 }
	if valY < 0 { valY = 0 }

	denom := math.Sqrt(valX * valY)
	if denom == 0 {
		return 0
	}
	return (float64(n)*sumXY - sumX*sumY) / denom
}

func (h *AnalyticsHandler) GetPerformance(c *gin.Context) {
	// Query the latest successful backtest run for each alpha, joined with alpha and author info
	query := `
		WITH LatestBacktests AS (
			SELECT DISTINCT ON (alpha_id) 
				id, alpha_id, metrics, status, created_at
			FROM backtest_runs
			WHERE status = 'completed'
			ORDER BY alpha_id, created_at DESC
		)
		SELECT 
			lb.alpha_id,
			a.name as alpha_name,
			u.username as author_name,
			lb.metrics->>'total_pnl' as total_return,
			lb.metrics->>'sharpe_ratio' as sharpe,
			lb.metrics->>'win_rate' as win_rate,
			lb.metrics->>'max_drawdown' as max_drawdown,
			lb.metrics->'pnl_curve' as pnl_curve,
			lb.status
		FROM LatestBacktests lb
		JOIN alphas a ON lb.alpha_id = a.id
		JOIN users u ON a.author_id = u.id
		ORDER BY lb.created_at DESC
	`

	rows, err := h.db.GetDB().Query(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to aggregate performance"})
		return
	}
	defer rows.Close()

	var items []models.PerformanceItem
	for rows.Next() {
		var item models.PerformanceItem
		var totalReturn, sharpe, winRate, maxDrawdown *string // Use pointers for nullable/missing fields
		
		if err := rows.Scan(
			&item.AlphaID, &item.AlphaName, &item.AuthorName,
			&totalReturn, &sharpe, &winRate, &maxDrawdown,
			&item.PnLCurve, &item.Status,
		); err != nil {
			log.Printf("error scanning performance row: %v", err)
			continue
		}

		// Convert string metrics from JSONB to float64 safely
		if totalReturn != nil { fmt.Sscanf(*totalReturn, "%f", &item.TotalReturn) }
		if sharpe != nil { fmt.Sscanf(*sharpe, "%f", &item.Sharpe) }
		if winRate != nil { fmt.Sscanf(*winRate, "%f", &item.WinRate) }
		if maxDrawdown != nil { fmt.Sscanf(*maxDrawdown, "%f", &item.MaxDrawdown) }

		items = append(items, item)
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
