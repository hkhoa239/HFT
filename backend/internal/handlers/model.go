package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"quantalpha/internal/database"
	"quantalpha/internal/models"
	"quantalpha/internal/redis"
	"quantalpha/internal/validator"
)

type ModelHandler struct {
	db   *database.Queries
	prod *redis.Producer
}

func NewModelHandler(db *database.Queries, prod *redis.Producer) *ModelHandler {
	return &ModelHandler{db: db, prod: prod}
}

func (h *ModelHandler) ListModels(c *gin.Context) {
	modelList, err := h.db.ListModels(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to list models"})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: modelList})
}

func (h *ModelHandler) TrainModel(c *gin.Context) {
	var req models.TrainModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "invalid request body"})
		return
	}
	if errs := validator.Validate(req); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: validator.GetFirstError(errs)})
		return
	}

	userID, _ := c.Get("user_id")
	model, err := h.db.CreateModel(c.Request.Context(), req.Name, req.Version, userID.(uuid.UUID), "", req.Params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to create model training job"})
		return
	}

	payload := redis.NewJobPayload("train", userID.(uuid.UUID).String(), "", map[string]interface{}{
		"model_name": req.Name,
		"version":    req.Version,
		"params":     req.Params,
	})
	payload.JobID = model.ID.String()
	if err := h.prod.PublishJob(c.Request.Context(), payload); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to queue training job"})
		return
	}

	user, _ := c.Get("user")
	h.db.CreateAuditLog(c.Request.Context(), user.(*models.User).ID, models.AuditCreate, models.EntityModel, model.ID, "create model training")

	c.JSON(http.StatusCreated, models.APIResponse{Success: true, Data: model})
}

func (h *ModelHandler) DeleteModel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "invalid model ID"})
		return
	}

	if err := h.db.DeleteModel(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to delete model"})
		return
	}

	user, _ := c.Get("user")
	h.db.CreateAuditLog(c.Request.Context(), user.(*models.User).ID, models.AuditDelete, models.EntityModel, id, "delete model")

	c.JSON(http.StatusOK, models.APIResponse{Success: true})
}
