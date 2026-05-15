package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"quantalpha/internal/database"
	"quantalpha/internal/models"
	"quantalpha/internal/validator"
)

type AlphaHandler struct {
	db *database.Queries
}

func NewAlphaHandler(db *database.Queries) *AlphaHandler {
	return &AlphaHandler{db: db}
}

func (h *AlphaHandler) ListMyAlphas(c *gin.Context) {
	userID, _ := c.Get("user_id")
	alphas, err := h.db.ListAlphasByAuthor(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to list alphas"})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: alphas})
}

func (h *AlphaHandler) ListMySubmittedAlphas(c *gin.Context) {
	userID, _ := c.Get("user_id")
	alphas, err := h.db.ListAlphasByAuthor(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to list alphas"})
		return
	}
	var submitted []models.Alpha
	for _, a := range alphas {
		if a.Status == models.AlphaStatusSubmitted {
			submitted = append(submitted, a)
		}
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: submitted})
}

func (h *AlphaHandler) CreateAlpha(c *gin.Context) {
	var req models.CreateAlphaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "invalid request body"})
		return
	}
	if errs := validator.Validate(req); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: validator.GetFirstError(errs)})
		return
	}
	userID, _ := c.Get("user_id")
	alpha, err := h.db.CreateAlpha(c.Request.Context(), userID.(uuid.UUID), req.Name, req.Description, req.CodeContent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to create alpha"})
		return
	}
	user, _ := c.Get("user")
	h.db.CreateAuditLog(c.Request.Context(), user.(*models.User).ID, models.AuditCreate, models.EntityAlpha, alpha.ID, "create alpha")
	c.JSON(http.StatusCreated, models.APIResponse{Success: true, Data: alpha})
}

func (h *AlphaHandler) UpdateAlpha(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "invalid alpha ID"})
		return
	}
	var req models.UpdateAlphaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "invalid request body"})
		return
	}
	if errs := validator.Validate(req); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: validator.GetFirstError(errs)})
		return
	}
	userID, _ := c.Get("user_id")
	alpha, err := h.db.GetAlphaByID(c.Request.Context(), id)
	if err != nil || alpha.AuthorID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, models.APIResponse{Success: false, Error: "not authorized"})
		return
	}
	updated, err := h.db.UpdateAlpha(c.Request.Context(), id, req.Name, req.Description, req.CodeContent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to update alpha"})
		return
	}
	user, _ := c.Get("user")
	h.db.CreateAuditLog(c.Request.Context(), user.(*models.User).ID, models.AuditUpdate, models.EntityAlpha, id, "update alpha")
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: updated})
}

func (h *AlphaHandler) SubmitAlpha(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "invalid alpha ID"})
		return
	}
	userID, _ := c.Get("user_id")
	alpha, err := h.db.GetAlphaByID(c.Request.Context(), id)
	if err != nil || alpha.AuthorID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, models.APIResponse{Success: false, Error: "not authorized"})
		return
	}
	submitted, err := h.db.SubmitAlpha(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to submit alpha"})
		return
	}
	user, _ := c.Get("user")
	h.db.CreateAuditLog(c.Request.Context(), user.(*models.User).ID, models.AuditSubmit, models.EntityAlpha, id, "submit alpha")
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: submitted})
}

func (h *AlphaHandler) DeleteAlpha(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "invalid alpha ID"})
		return
	}
	userID, _ := c.Get("user_id")
	alpha, err := h.db.GetAlphaByID(c.Request.Context(), id)
	if err != nil || alpha.AuthorID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, models.APIResponse{Success: false, Error: "not authorized"})
		return
	}
	if err := h.db.DeleteAlpha(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to delete alpha"})
		return
	}
	user, _ := c.Get("user")
	h.db.CreateAuditLog(c.Request.Context(), user.(*models.User).ID, models.AuditDelete, models.EntityAlpha, id, "delete alpha")
	c.JSON(http.StatusOK, models.APIResponse{Success: true})
}

func (h *AlphaHandler) ListSubmittedAlphas(c *gin.Context) {
	alphas, err := h.db.ListSubmittedAlphas(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to list submitted alphas"})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: alphas})
}
