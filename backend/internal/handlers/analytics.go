package handlers

import (
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
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: models.PerformanceResponse{
		Data: []models.PerformanceItem{},
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
