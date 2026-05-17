package database

import (
	"context"
	"github.com/google/uuid"
	"quantalpha/internal/models"
)

type Querier interface {
	// User methods
	GetUserByUsername(ctx context.Context, username string) (*models.User, string, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	ListUsers(ctx context.Context) ([]models.User, error)
	CreateUser(ctx context.Context, username, passwordHash string, role models.UserRole, fullName string) (*models.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, username, fullName string, role models.UserRole) (*models.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error

	// Factor methods
	CreateFactor(ctx context.Context, name, description, dataPath, frequency string, createdBy uuid.UUID) (*models.Factor, error)
	GetFactorByID(ctx context.Context, id uuid.UUID) (*models.Factor, error)
	ListFactors(ctx context.Context) ([]models.Factor, error)
	UpdateFactor(ctx context.Context, id uuid.UUID, name, description, dataPath, frequency string) (*models.Factor, error)
	DeleteFactor(ctx context.Context, id uuid.UUID) error

	// Alpha methods
	CreateAlpha(ctx context.Context, authorID uuid.UUID, name, description, codeContent string) (*models.Alpha, error)
	GetAlphaByID(ctx context.Context, id uuid.UUID) (*models.Alpha, error)
	ListAlphasByAuthor(ctx context.Context, authorID uuid.UUID) ([]models.Alpha, error)
	ListSubmittedAlphas(ctx context.Context) ([]models.Alpha, error)
	UpdateAlpha(ctx context.Context, id uuid.UUID, name, description, codeContent string) (*models.Alpha, error)
	SubmitAlpha(ctx context.Context, id uuid.UUID) (*models.Alpha, error)
	DeleteAlpha(ctx context.Context, id uuid.UUID) error

	// Backtest methods
	CreateBacktestRun(ctx context.Context, alphaID, executorID uuid.UUID, params map[string]interface{}) (*models.BacktestRun, error)
	GetBacktestRunByID(ctx context.Context, id uuid.UUID) (*models.BacktestRun, error)
	ListBacktestRunsByExecutor(ctx context.Context, executorID uuid.UUID) ([]models.BacktestRun, error)
	ListBacktestRuns(ctx context.Context) ([]models.BacktestRun, error)
	UpdateBacktestRunStatus(ctx context.Context, id uuid.UUID, status models.JobStatus, metrics map[string]interface{}, errorLog string) error
	DeleteJob(ctx context.Context, jobID uuid.UUID) error
	FindActiveBacktestRun(ctx context.Context, alphaID uuid.UUID, params map[string]interface{}) (*models.BacktestRun, error)

	// Model methods
	CreateModel(ctx context.Context, name, version string, dsID uuid.UUID, pklPath string, trainingParams map[string]interface{}) (*models.Model, error)
	GetModelByID(ctx context.Context, id uuid.UUID) (*models.Model, error)
	ListModels(ctx context.Context) ([]models.Model, error)
	DeleteModel(ctx context.Context, id uuid.UUID) error

	// Audit methods
	CreateAuditLog(ctx context.Context, userID uuid.UUID, action models.AuditAction, entityType models.AuditEntityType, entityID uuid.UUID, details string) error
	ListAuditLogs(ctx context.Context, limit, offset int) ([]models.AuditLog, int, error)

	// Analytics methods
	GetPerformance(ctx context.Context) ([]models.PerformanceItem, error)
	GetCorrelationData(ctx context.Context) ([]models.CorrelationItemRaw, error)

	// Health check
	Ping(ctx context.Context) error
}
