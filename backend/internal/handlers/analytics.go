package handlers

import (
	"fmt"
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
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: models.CorrelationResponse{
		Data: []models.CorrelationItem{},
	}})
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
			continue
		}

		// Convert string metrics from JSONB to float64 safely
		if totalReturn != nil { fmt.Sscanf(*totalReturn, "%f", &item.TotalReturn) }
		if sharpe != nil { fmt.Sscanf(*sharpe, "%f", &item.Sharpe) }
		if winRate != nil { fmt.Sscanf(*winRate, "%f", &item.WinRate) }
		if maxDrawdown != nil { fmt.Sscanf(*maxDrawdown, "%f", &item.MaxDrawdown) }

		items = append(items, item)
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: models.PerformanceResponse{
		Data: items,
	}})
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
