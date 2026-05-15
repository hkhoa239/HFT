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

type BacktestHandler struct {
	db   *database.Queries
	prod *redis.Producer
}

func NewBacktestHandler(db *database.Queries, prod *redis.Producer) *BacktestHandler {
	return &BacktestHandler{db: db, prod: prod}
}

func (h *BacktestHandler) RunBacktest(c *gin.Context) {
	var req models.RunBacktestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "invalid request body"})
		return
	}
	if errs := validator.Validate(req); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: validator.GetFirstError(errs)})
		return
	}

	alphaUUID, _ := uuid.Parse(req.AlphaID)
	alpha, err := h.db.GetAlphaByID(c.Request.Context(), alphaUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Error: "alpha not found"})
		return
	}

	userID, _ := c.Get("user_id")
	backtest, err := h.db.CreateBacktestRun(c.Request.Context(), alphaUUID, userID.(uuid.UUID), map[string]interface{}{
		"start":   req.Params.Start,
		"end":     req.Params.End,
		"capital": req.Params.Capital,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to create backtest"})
		return
	}

	payload := redis.NewJobPayload("backtest", userID.(uuid.UUID).String(), req.AlphaID, alpha.CodeContent, map[string]interface{}{
		"start":   req.Params.Start,
		"end":     req.Params.End,
		"capital": req.Params.Capital,
	})
	payload.JobID = backtest.ID.String()

	if h.prod == nil {
		c.JSON(http.StatusServiceUnavailable, models.APIResponse{Success: false, Error: "backtest engine is currently unavailable (redis offline)"})
		return
	}

	if err := h.prod.PublishJob(c.Request.Context(), payload); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to queue job"})
		return
	}

	user, _ := c.Get("user")
	h.db.CreateAuditLog(c.Request.Context(), user.(*models.User).ID, models.AuditRun, models.EntityBacktest, backtest.ID, "run backtest")

	c.JSON(http.StatusCreated, models.APIResponse{Success: true, Data: backtest})
}

func (h *BacktestHandler) GetMyBacktestStatus(c *gin.Context) {
	userID, _ := c.Get("user_id")
	runs, err := h.db.ListBacktestRunsByExecutor(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to list backtests"})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: runs})
}

func (h *BacktestHandler) GetBacktestStatus(c *gin.Context) {
	idStr := c.Param("job_id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "invalid job ID"})
		return
	}

	run, err := h.db.GetBacktestRunByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Error: "backtest not found"})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: run})
}

func (h *BacktestHandler) ListAllBacktestStatus(c *gin.Context) {
	runs, err := h.db.ListBacktestRuns(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to list backtests"})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: runs})
}
