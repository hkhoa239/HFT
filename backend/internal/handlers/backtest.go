package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"quantalpha/internal/database"
	"quantalpha/internal/models"
	"quantalpha/internal/redis"
	"quantalpha/internal/validator"
	"sync"
)

type BacktestHandler struct {
	db   database.Querier
	prod redis.JobProducer
	mu   sync.Mutex
}

func NewBacktestHandler(db database.Querier, prod redis.JobProducer) *BacktestHandler {
	return &BacktestHandler{db: db, prod: prod}
}

func (h *BacktestHandler) RunBacktest(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

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

	instrument := req.Params.Instrument
	if instrument == "" {
		instrument = "VN30F2112"
	}
	lookbackSec := req.Params.LookbackSec
	if lookbackSec <= 0 {
		lookbackSec = 60
	}
	predictionSec := req.Params.PredictionSec
	if predictionSec <= 0 {
		predictionSec = 10
	}

	params := map[string]interface{}{
		"start":          req.Params.Start,
		"end":            req.Params.End,
		"capital":        req.Params.Capital,
		"instrument":     instrument,
		"lookback_sec":   lookbackSec,
		"prediction_sec": predictionSec,
	}

	// Idempotency check: Don't run if already pending or running with same params
	existing, _ := h.db.FindActiveBacktestRun(c.Request.Context(), alphaUUID, params)
	if existing != nil {
		c.JSON(http.StatusConflict, models.APIResponse{Success: false, Error: "A backtest with these parameters is already in progress"})
		return
	}

	alpha, err := h.db.GetAlphaByID(c.Request.Context(), alphaUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Error: "alpha not found"})
		return
	}

	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uuid.UUID)

	// Verify executor exists in DB (safety check for local dev DB resets)
	executor, err := h.db.GetUserByID(c.Request.Context(), userID)
	if err != nil || executor == nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Error: "authenticated user not found in database"})
		return
	}

	backtest, err := h.db.CreateBacktestRun(c.Request.Context(), alphaUUID, userID, params)
	if err != nil {
		log.Printf("Error creating backtest run: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to create backtest: " + err.Error()})
		return
	}

	payload := redis.NewJobPayload("backtest", userID.String(), req.AlphaID, alpha.CodeContent, map[string]interface{}{
		"start":          req.Params.Start,
		"end":            req.Params.End,
		"capital":        req.Params.Capital,
		"instrument":     instrument,
		"lookback_sec":   lookbackSec,
		"prediction_sec": predictionSec,
	})
	payload.JobID = backtest.ID.String()

	if h.prod == nil {
		h.db.DeleteJob(c.Request.Context(), backtest.ID)
		c.JSON(http.StatusServiceUnavailable, models.APIResponse{Success: false, Error: "backtest engine is currently unavailable (redis offline)"})
		return
	}

	if err := h.prod.PublishJob(c.Request.Context(), payload); err != nil {
		h.db.DeleteJob(c.Request.Context(), backtest.ID)
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to queue job"})
		return
	}

	user, _ := c.Get("user")
	if err := h.db.CreateAuditLog(c.Request.Context(), user.(*models.User).ID, models.AuditRun, models.EntityBacktest, backtest.ID, "run backtest"); err != nil {
		// Log error but don't fail the request since the job is already in Redis
		// In a real system, we might use a background worker for audit logs or retry
		// fmt.Printf("Failed to create audit log: %v\n", err)
	}

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
