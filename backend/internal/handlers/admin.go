package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"quantalpha/internal/database"
	"quantalpha/internal/models"
)

type AdminHandler struct {
	db *database.Queries
}

func NewAdminHandler(db *database.Queries) *AdminHandler {
	return &AdminHandler{db: db}
}

func (h *AdminHandler) DeleteJob(c *gin.Context) {
	jobIDStr := c.Param("job_id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "invalid job ID"})
		return
	}

	if err := h.db.DeleteJob(c.Request.Context(), jobID); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to delete job"})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true})
}
