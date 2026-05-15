package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleQR    UserRole = "qr"
	RolePM    UserRole = "pm"
	RoleDS    UserRole = "ds"
)

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

type AlphaStatus string

const (
	AlphaStatusDraft     AlphaStatus = "draft"
	AlphaStatusSubmitted AlphaStatus = "submitted"
)

type AuditAction string

const (
	AuditCreate AuditAction = "create"
	AuditUpdate AuditAction = "update"
	AuditDelete AuditAction = "delete"
	AuditRun    AuditAction = "run"
	AuditSubmit AuditAction = "submit"
)

type AuditEntityType string

const (
	EntityUser     AuditEntityType = "user"
	EntityAlpha    AuditEntityType = "alpha"
	EntityBacktest AuditEntityType = "backtest"
	EntityModel    AuditEntityType = "model"
	EntityFactor   AuditEntityType = "factor"
)

type User struct {
	ID            uuid.UUID `json:"id"`
	Username      string    `json:"username"`
	FullName      string    `json:"full_name,omitempty"`
	Role          UserRole  `json:"role"`
	DummyPassword string    `json:"-"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateUserRequest struct {
	Username string   `json:"username" validate:"required,min=3,max=50"`
	Password string   `json:"password" validate:"required,min=6"`
	FullName string   `json:"full_name" validate:"max=100"`
	Role     UserRole `json:"role" validate:"required,oneof=admin qr pm ds"`
}

type UpdateUserRequest struct {
	Username string   `json:"username" validate:"min=3,max=50"`
	FullName string   `json:"full_name" validate:"max=100"`
	Role     UserRole `json:"role" validate:"oneof=admin qr pm ds"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

type Factor struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	DataPath    string    `json:"data_path"`
	Frequency   string    `json:"frequency"`
	CreatedBy   uuid.UUID `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateFactorRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description"`
	DataPath    string `json:"data_path" validate:"required"`
	Frequency   string `json:"frequency" validate:"oneof=1m 5m 15m 1h 1d"`
}

type UpdateFactorRequest struct {
	Name        string `json:"name" validate:"min=1,max=100"`
	Description string `json:"description"`
	DataPath    string `json:"data_path"`
	Frequency   string `json:"frequency" validate:"oneof=1m 5m 15m 1h 1d"`
}

type FactorPreviewResponse struct {
	Data []map[string]interface{} `json:"data"`
}

type Alpha struct {
	ID          uuid.UUID   `json:"id"`
	AuthorID    uuid.UUID   `json:"author_id"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	CodeContent string      `json:"code_content"`
	Status      AlphaStatus `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type CreateAlphaRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description"`
	CodeContent string `json:"code_content" validate:"required"`
}

type UpdateAlphaRequest struct {
	Name        string `json:"name" validate:"min=1,max=100"`
	Description string `json:"description"`
	CodeContent string `json:"code_content"`
}

type BacktestRun struct {
	ID         uuid.UUID              `json:"id"`
	AlphaID    uuid.UUID              `json:"alpha_id"`
	ExecutorID uuid.UUID              `json:"executor_id"`
	Status     JobStatus              `json:"status"`
	Params     map[string]interface{} `json:"params"`
	Metrics    map[string]interface{} `json:"metrics,omitempty"`
	ErrorLog   string                 `json:"error_log,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	FinishedAt *time.Time             `json:"finished_at,omitempty"`
}

type BacktestParams struct {
	Start   string  `json:"start" validate:"required"`
	End     string  `json:"end" validate:"required"`
	Capital float64 `json:"capital" validate:"required,gt=0"`
}

type RunBacktestRequest struct {
	AlphaID string          `json:"alpha_id" validate:"required,uuid"`
	Params  *BacktestParams `json:"params" validate:"required"`
}

type BacktestStatusResponse struct {
	Status   JobStatus              `json:"status"`
	Metrics  map[string]interface{} `json:"metrics,omitempty"`
	ErrorLog string                 `json:"error_log,omitempty"`
}

type Model struct {
	ID              uuid.UUID              `json:"id"`
	Name            string                 `json:"name"`
	Version         string                 `json:"version"`
	DSID            uuid.UUID              `json:"ds_id"`
	PklPath         string                 `json:"pkl_path"`
	TrainingMetrics map[string]interface{} `json:"training_metrics,omitempty"`
	TrainingParams  map[string]interface{} `json:"training_params,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

type TrainModelRequest struct {
	Name    string                 `json:"name" validate:"required,min=1,max=100"`
	Version string                 `json:"version" validate:"required"`
	Params  map[string]interface{} `json:"params"`
}

type ModelPreviewResponse struct {
	Data []map[string]interface{} `json:"data"`
}

type CorrelationItem struct {
	AlphaID1 uuid.UUID `json:"alpha_id_1"`
	AlphaID2 uuid.UUID `json:"alpha_id_2"`
	Value    float64   `json:"value"`
}

type CorrelationResponse struct {
	Data []CorrelationItem `json:"data"`
}

type PerformanceItem struct {
	AlphaID     uuid.UUID `json:"alpha_id"`
	TotalReturn float64   `json:"total_return"`
	Sharpe      float64   `json:"sharpe"`
	MaxDrawdown float64   `json:"max_drawdown"`
}

type PerformanceResponse struct {
	Data []PerformanceItem `json:"data"`
}

type AuditLog struct {
	ID         int             `json:"id"`
	UserID     uuid.UUID       `json:"user_id"`
	Action     AuditAction     `json:"action"`
	EntityType AuditEntityType `json:"entity_type"`
	EntityID   uuid.UUID       `json:"entity_id,omitempty"`
	Details    string          `json:"details,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

type AuditLogResponse struct {
	Data  []AuditLog `json:"data"`
	Total int        `json:"total"`
}

type JobPayload struct {
	JobID     string                 `json:"job_id"`
	TaskType  string                 `json:"task_type"`
	UserID    string                 `json:"user_id"`
	AlphaID   string                 `json:"alpha_id,omitempty"`
	Params    map[string]interface{} `json:"params"`
	CreatedAt string                 `json:"created_at"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
