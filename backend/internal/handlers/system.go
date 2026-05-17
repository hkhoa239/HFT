package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"quantalpha/internal/database"
	"quantalpha/internal/models"
	"quantalpha/internal/redis"
)

type SystemHandler struct {
	db   database.Querier
	prod redis.JobProducer
}

func NewSystemHandler(db database.Querier, prod redis.JobProducer) *SystemHandler {
	return &SystemHandler{db: db, prod: prod}
}

func (h *SystemHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: "ok"})
}

func (h *SystemHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	// Check DB
	if err := h.db.Ping(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, models.APIResponse{Success: false, Error: "database unreachable"})
		return
	}

	// Check Redis (if producer is configured)
	if h.prod == nil {
		// Redis is optional for some routes, but we report it as not ready if critical for MVP
		c.JSON(http.StatusServiceUnavailable, models.APIResponse{Success: false, Error: "redis producer not initialized"})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: "ready"})
}
