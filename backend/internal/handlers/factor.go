package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"quantalpha/internal/database"
	"quantalpha/internal/models"
	"quantalpha/internal/validator"
)

type FactorHandler struct {
	db *database.Queries
}

func NewFactorHandler(db *database.Queries) *FactorHandler {
	return &FactorHandler{db: db}
}

func (h *FactorHandler) ListFactors(c *gin.Context) {
	factors, err := h.db.ListFactors(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to list factors"})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: factors})
}

func (h *FactorHandler) CreateFactor(c *gin.Context) {
	var req models.CreateFactorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "invalid request body"})
		return
	}
	if errs := validator.Validate(req); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: validator.GetFirstError(errs)})
		return
	}
	userID, _ := c.Get("user_id")
	factor, err := h.db.CreateFactor(c.Request.Context(), req.Name, req.Description, req.DataPath, req.Frequency, userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to create factor"})
		return
	}
	user, _ := c.Get("user")
	h.db.CreateAuditLog(c.Request.Context(), user.(*models.User).ID, models.AuditCreate, models.EntityFactor, factor.ID, "create factor")
	c.JSON(http.StatusCreated, models.APIResponse{Success: true, Data: factor})
}

func (h *FactorHandler) UpdateFactor(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "invalid factor ID"})
		return
	}
	var req models.UpdateFactorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "invalid request body"})
		return
	}
	if errs := validator.Validate(req); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: validator.GetFirstError(errs)})
		return
	}
	updated, err := h.db.UpdateFactor(c.Request.Context(), id, req.Name, req.Description, req.DataPath, req.Frequency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to update factor"})
		return
	}
	user, _ := c.Get("user")
	h.db.CreateAuditLog(c.Request.Context(), user.(*models.User).ID, models.AuditUpdate, models.EntityFactor, id, "update factor")
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: updated})
}

func (h *FactorHandler) DeleteFactor(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "invalid factor ID"})
		return
	}
	if err := h.db.DeleteFactor(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to delete factor"})
		return
	}
	user, _ := c.Get("user")
	h.db.CreateAuditLog(c.Request.Context(), user.(*models.User).ID, models.AuditDelete, models.EntityFactor, id, "delete factor")
	c.JSON(http.StatusOK, models.APIResponse{Success: true})
}

func (h *FactorHandler) GetFactorPreview(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "invalid factor ID"})
		return
	}
	factor, err := h.db.GetFactorByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Error: "factor not found"})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: models.FactorPreviewResponse{
		Data: []map[string]interface{}{
			{"message": "preview not available - worker must create _preview.json"},
		},
	}})
	_ = factor
}
