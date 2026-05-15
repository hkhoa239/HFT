package database

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"quantalpha/internal/models"
)

type Queries struct {
	db *pgxpool.Pool
}

func NewQueries(db *pgxpool.Pool) *Queries {
	return &Queries{db: db}
}

func (q *Queries) CreateUser(ctx context.Context, username, passwordHash string, role models.UserRole, fullName string) (*models.User, error) {
	user := &models.User{
		ID:       uuid.New(),
		Username: username,
		Role:     role,
		FullName: fullName,
	}

	if fullName != "" {
		user.FullName = fullName
	}

	query := `
		INSERT INTO users (id, username, password_hash, role, full_name)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, username, role, full_name, created_at, updated_at
	`

	err := q.db.QueryRow(ctx, query, user.ID, user.Username, passwordHash, user.Role, user.FullName).Scan(
		&user.ID, &user.Username, &user.Role, &user.FullName, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (q *Queries) GetUserByUsername(ctx context.Context, username string) (*models.User, string, error) {
	var user models.User
	var passwordHash string

	query := `
		SELECT id, username, password_hash, role, full_name, created_at, updated_at
		FROM users
		WHERE username = $1 AND deleted_at IS NULL
	`

	err := q.db.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &passwordHash, &user.Role, &user.FullName, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, "", fmt.Errorf("user not found: %w", err)
	}

	return &user, passwordHash, nil
}

func (q *Queries) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User

	query := `
		SELECT id, username, role, full_name, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := q.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Role, &user.FullName, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

func (q *Queries) ListUsers(ctx context.Context) ([]models.User, error) {
	query := `
		SELECT id, username, role, full_name, created_at, updated_at
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := q.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Role, &user.FullName, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (q *Queries) UpdateUser(ctx context.Context, id uuid.UUID, username, fullName string, role models.UserRole) (*models.User, error) {
	var user models.User

	query := `
		UPDATE users
		SET username = COALESCE(NULLIF($2, ''), full_name = COALESCE(NULLIF($3, '')), role = $4
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, username, role, full_name, created_at, updated_at
	`

	err := q.db.QueryRow(ctx, query, id, username, fullName, role).Scan(
		&user.ID, &user.Username, &user.Role, &user.FullName, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &user, nil
}

func (q *Queries) DeleteUser(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND deleted_at IS NULL`
	_, err := q.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (q *Queries) CreateFactor(ctx context.Context, name, description, dataPath, frequency string, createdBy uuid.UUID) (*models.Factor, error) {
	factor := &models.Factor{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		DataPath:    dataPath,
		Frequency:   frequency,
		CreatedBy:   createdBy,
	}

	query := `
		INSERT INTO factors (id, name, description, data_path, frequency, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, description, data_path, frequency, created_by, created_at, updated_at
	`

	err := q.db.QueryRow(ctx, query, factor.ID, factor.Name, factor.Description, factor.DataPath, factor.Frequency, factor.CreatedBy).Scan(
		&factor.ID, &factor.Name, &factor.Description, &factor.DataPath, &factor.Frequency, &factor.CreatedBy, &factor.CreatedAt, &factor.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create factor: %w", err)
	}

	return factor, nil
}

func (q *Queries) GetFactorByID(ctx context.Context, id uuid.UUID) (*models.Factor, error) {
	var factor models.Factor

	query := `
		SELECT id, name, description, data_path, frequency, created_by, created_at, updated_at
		FROM factors
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := q.db.QueryRow(ctx, query, id).Scan(
		&factor.ID, &factor.Name, &factor.Description, &factor.DataPath, &factor.Frequency, &factor.CreatedBy, &factor.CreatedAt, &factor.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("factor not found: %w", err)
	}

	return &factor, nil
}

func (q *Queries) ListFactors(ctx context.Context) ([]models.Factor, error) {
	query := `
		SELECT id, name, description, data_path, frequency, created_by, created_at, updated_at
		FROM factors
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := q.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list factors: %w", err)
	}
	defer rows.Close()

	var factors []models.Factor
	for rows.Next() {
		var factor models.Factor
		if err := rows.Scan(&factor.ID, &factor.Name, &factor.Description, &factor.DataPath, &factor.Frequency, &factor.CreatedBy, &factor.CreatedAt, &factor.UpdatedAt); err != nil {
			return nil, err
		}
		factors = append(factors, factor)
	}

	return factors, nil
}

func (q *Queries) UpdateFactor(ctx context.Context, id uuid.UUID, name, description, dataPath, frequency string) (*models.Factor, error) {
	var factor models.Factor

	query := `
		UPDATE factors
		SET name = COALESCE(NULLIF($2, ''), description = COALESCE(NULLIF($3, '')), data_path = COALESCE(NULLIF($4, '')), frequency = COALESCE(NULLIF($5, ''))
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, name, description, data_path, frequency, created_by, created_at, updated_at
	`

	err := q.db.QueryRow(ctx, query, id, name, description, dataPath, frequency).Scan(
		&factor.ID, &factor.Name, &factor.Description, &factor.DataPath, &factor.Frequency, &factor.CreatedBy, &factor.CreatedAt, &factor.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update factor: %w", err)
	}

	return &factor, nil
}

func (q *Queries) DeleteFactor(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE factors SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND deleted_at IS NULL`
	_, err := q.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete factor: %w", err)
	}
	return nil
}

func (q *Queries) CreateAlpha(ctx context.Context, authorID uuid.UUID, name, description, codeContent string) (*models.Alpha, error) {
	alpha := &models.Alpha{
		ID:          uuid.New(),
		AuthorID:    authorID,
		Name:        name,
		Description: description,
		CodeContent: codeContent,
		Status:      models.AlphaStatusDraft,
	}

	query := `
		INSERT INTO alphas (id, author_id, name, description, code_content, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, author_id, name, description, code_content, status, created_at, updated_at
	`

	err := q.db.QueryRow(ctx, query, alpha.ID, alpha.AuthorID, alpha.Name, alpha.Description, alpha.CodeContent, alpha.Status).Scan(
		&alpha.ID, &alpha.AuthorID, &alpha.Name, &alpha.Description, &alpha.CodeContent, &alpha.Status, &alpha.CreatedAt, &alpha.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create alpha: %w", err)
	}

	return alpha, nil
}

func (q *Queries) GetAlphaByID(ctx context.Context, id uuid.UUID) (*models.Alpha, error) {
	var alpha models.Alpha

	query := `
		SELECT id, author_id, name, description, code_content, status, created_at, updated_at
		FROM alphas
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := q.db.QueryRow(ctx, query, id).Scan(
		&alpha.ID, &alpha.AuthorID, &alpha.Name, &alpha.Description, &alpha.CodeContent, &alpha.Status, &alpha.CreatedAt, &alpha.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("alpha not found: %w", err)
	}

	return &alpha, nil
}

func (q *Queries) ListAlphasByAuthor(ctx context.Context, authorID uuid.UUID) ([]models.Alpha, error) {
	query := `
		SELECT id, author_id, name, description, code_content, status, created_at, updated_at
		FROM alphas
		WHERE author_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := q.db.Query(ctx, query, authorID)
	if err != nil {
		return nil, fmt.Errorf("failed to list alphas: %w", err)
	}
	defer rows.Close()

	var alphas []models.Alpha
	for rows.Next() {
		var alpha models.Alpha
		if err := rows.Scan(&alpha.ID, &alpha.AuthorID, &alpha.Name, &alpha.Description, &alpha.CodeContent, &alpha.Status, &alpha.CreatedAt, &alpha.UpdatedAt); err != nil {
			return nil, err
		}
		alphas = append(alphas, alpha)
	}

	return alphas, nil
}

func (q *Queries) ListSubmittedAlphas(ctx context.Context) ([]models.Alpha, error) {
	query := `
		SELECT id, author_id, name, description, code_content, status, created_at, updated_at
		FROM alphas
		WHERE status = 'submitted' AND deleted_at IS NULL
		ORDER BY updated_at DESC
	`

	rows, err := q.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list submitted alphas: %w", err)
	}
	defer rows.Close()

	var alphas []models.Alpha
	for rows.Next() {
		var alpha models.Alpha
		if err := rows.Scan(&alpha.ID, &alpha.AuthorID, &alpha.Name, &alpha.Description, &alpha.CodeContent, &alpha.Status, &alpha.CreatedAt, &alpha.UpdatedAt); err != nil {
			return nil, err
		}
		alphas = append(alphas, alpha)
	}

	return alphas, nil
}

func (q *Queries) UpdateAlpha(ctx context.Context, id uuid.UUID, name, description, codeContent string) (*models.Alpha, error) {
	var alpha models.Alpha

	query := `
		UPDATE alphas
		SET name = COALESCE(NULLIF($2, '')), description = COALESCE(NULLIF($3, '')), code_content = COALESCE(NULLIF($4, ''))
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, author_id, name, description, code_content, status, created_at, updated_at
	`

	err := q.db.QueryRow(ctx, query, id, name, description, codeContent).Scan(
		&alpha.ID, &alpha.AuthorID, &alpha.Name, &alpha.Description, &alpha.CodeContent, &alpha.Status, &alpha.CreatedAt, &alpha.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update alpha: %w", err)
	}

	return &alpha, nil
}

func (q *Queries) SubmitAlpha(ctx context.Context, id uuid.UUID) (*models.Alpha, error) {
	var alpha models.Alpha

	query := `
		UPDATE alphas
		SET status = 'submitted'
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, author_id, name, description, code_content, status, created_at, updated_at
	`

	err := q.db.QueryRow(ctx, query, id).Scan(
		&alpha.ID, &alpha.AuthorID, &alpha.Name, &alpha.Description, &alpha.CodeContent, &alpha.Status, &alpha.CreatedAt, &alpha.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to submit alpha: %w", err)
	}

	return &alpha, nil
}

func (q *Queries) DeleteAlpha(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE alphas SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND deleted_at IS NULL`
	_, err := q.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete alpha: %w", err)
	}
	return nil
}

func (q *Queries) CreateBacktestRun(ctx context.Context, alphaID, executorID uuid.UUID, params map[string]interface{}) (*models.BacktestRun, error) {
	backtest := &models.BacktestRun{
		ID:         uuid.New(),
		AlphaID:    alphaID,
		ExecutorID: executorID,
		Status:     models.JobStatusPending,
		Params:     params,
	}

	query := `
		INSERT INTO backtest_runs (id, alpha_id, executor_id, status, params)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, alpha_id, executor_id, status, params, metrics, error_log, created_at, finished_at
	`

	err := q.db.QueryRow(ctx, query, backtest.ID, backtest.AlphaID, backtest.ExecutorID, backtest.Status, backtest.Params).Scan(
		&backtest.ID, &backtest.AlphaID, &backtest.ExecutorID, &backtest.Status, &backtest.Params, &backtest.Metrics, &backtest.ErrorLog, &backtest.CreatedAt, &backtest.FinishedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create backtest run: %w", err)
	}

	return backtest, nil
}

func (q *Queries) GetBacktestRunByID(ctx context.Context, id uuid.UUID) (*models.BacktestRun, error) {
	var backtest models.BacktestRun

	query := `
		SELECT id, alpha_id, executor_id, status, params, metrics, error_log, created_at, finished_at
		FROM backtest_runs
		WHERE id = $1
	`

	err := q.db.QueryRow(ctx, query, id).Scan(
		&backtest.ID, &backtest.AlphaID, &backtest.ExecutorID, &backtest.Status, &backtest.Params, &backtest.Metrics, &backtest.ErrorLog, &backtest.CreatedAt, &backtest.FinishedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("backtest run not found: %w", err)
	}

	return &backtest, nil
}

func (q *Queries) ListBacktestRunsByExecutor(ctx context.Context, executorID uuid.UUID) ([]models.BacktestRun, error) {
	query := `
		SELECT id, alpha_id, executor_id, status, params, metrics, error_log, created_at, finished_at
		FROM backtest_runs
		WHERE executor_id = $1
		ORDER BY created_at DESC
	`

	rows, err := q.db.Query(ctx, query, executorID)
	if err != nil {
		return nil, fmt.Errorf("failed to list backtest runs: %w", err)
	}
	defer rows.Close()

	var runs []models.BacktestRun
	for rows.Next() {
		var run models.BacktestRun
		if err := rows.Scan(&run.ID, &run.AlphaID, &run.ExecutorID, &run.Status, &run.Params, &run.Metrics, &run.ErrorLog, &run.CreatedAt, &run.FinishedAt); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}

	return runs, nil
}

func (q *Queries) ListBacktestRuns(ctx context.Context) ([]models.BacktestRun, error) {
	query := `
		SELECT id, alpha_id, executor_id, status, params, metrics, error_log, created_at, finished_at
		FROM backtest_runs
		ORDER BY created_at DESC
	`

	rows, err := q.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list backtest runs: %w", err)
	}
	defer rows.Close()

	var runs []models.BacktestRun
	for rows.Next() {
		var run models.BacktestRun
		if err := rows.Scan(&run.ID, &run.AlphaID, &run.ExecutorID, &run.Status, &run.Params, &run.Metrics, &run.ErrorLog, &run.CreatedAt, &run.FinishedAt); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}

	return runs, nil
}

func (q *Queries) UpdateBacktestRunStatus(ctx context.Context, id uuid.UUID, status models.JobStatus, metrics map[string]interface{}, errorLog string) error {
	query := `
		UPDATE backtest_runs
		SET status = $2, metrics = $3, error_log = $4, finished_at = CASE WHEN $2 IN ('completed', 'failed') THEN CURRENT_TIMESTAMP ELSE finished_at END
		WHERE id = $1
	`

	_, err := q.db.Exec(ctx, query, id, status, metrics, errorLog)
	if err != nil {
		return fmt.Errorf("failed to update backtest run: %w", err)
	}
	return nil
}

func (q *Queries) CreateModel(ctx context.Context, name, version string, dsID uuid.UUID, pklPath string, trainingParams map[string]interface{}) (*models.Model, error) {
	model := &models.Model{
		ID:             uuid.New(),
		Name:           name,
		Version:        version,
		DSID:           dsID,
		PklPath:        pklPath,
		TrainingParams: trainingParams,
	}

	query := `
		INSERT INTO models (id, name, version, ds_id, pkl_path, training_params)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, version, ds_id, pkl_path, training_params, training_metrics, created_at, updated_at
	`

	err := q.db.QueryRow(ctx, query, model.ID, model.Name, model.Version, model.DSID, model.PklPath, model.TrainingParams).Scan(
		&model.ID, &model.Name, &model.Version, &model.DSID, &model.PklPath, &model.TrainingParams, &model.TrainingMetrics, &model.CreatedAt, &model.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create model: %w", err)
	}

	return model, nil
}

func (q *Queries) GetModelByID(ctx context.Context, id uuid.UUID) (*models.Model, error) {
	var model models.Model

	query := `
		SELECT id, name, version, ds_id, pkl_path, training_params, training_metrics, created_at, updated_at
		FROM models
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := q.db.QueryRow(ctx, query, id).Scan(
		&model.ID, &model.Name, &model.Version, &model.DSID, &model.PklPath, &model.TrainingParams, &model.TrainingMetrics, &model.CreatedAt, &model.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("model not found: %w", err)
	}

	return &model, nil
}

func (q *Queries) ListModels(ctx context.Context) ([]models.Model, error) {
	query := `
		SELECT id, name, version, ds_id, pkl_path, training_params, training_metrics, created_at, updated_at
		FROM models
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := q.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	defer rows.Close()

	var modelList []models.Model
	for rows.Next() {
		var model models.Model
		if err := rows.Scan(&model.ID, &model.Name, &model.Version, &model.DSID, &model.PklPath, &model.TrainingParams, &model.TrainingMetrics, &model.CreatedAt, &model.UpdatedAt); err != nil {
			return nil, err
		}
		modelList = append(modelList, model)
	}

	return modelList, nil
}

func (q *Queries) DeleteModel(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE models SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND deleted_at IS NULL`
	_, err := q.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete model: %w", err)
	}
	return nil
}

func (q *Queries) CreateAuditLog(ctx context.Context, userID uuid.UUID, action models.AuditAction, entityType models.AuditEntityType, entityID uuid.UUID, details string) error {
	query := `
		INSERT INTO audit_logs (user_id, action, entity_type, entity_id, details)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := q.db.Exec(ctx, query, userID, action, entityType, entityID, details)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	return nil
}

func (q *Queries) ListAuditLogs(ctx context.Context, limit, offset int) ([]models.AuditLog, int, error) {
	countQuery := `SELECT COUNT(*) FROM audit_logs`
	var total int
	if err := q.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	query := `
		SELECT id, user_id, action, entity_type, entity_id, details, created_at
		FROM audit_logs
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := q.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list audit logs: %w", err)
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		if err := rows.Scan(&log.ID, &log.UserID, &log.Action, &log.EntityType, &log.EntityID, &log.Details, &log.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}

	return logs, total, nil
}

func (q *Queries) DeleteJob(ctx context.Context, jobID uuid.UUID) error {
	query := `UPDATE backtest_runs SET status = 'failed', finished_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := q.db.Exec(ctx, query, jobID)
	if err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}
	return nil
}
