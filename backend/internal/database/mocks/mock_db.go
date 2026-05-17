package mocks

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"quantalpha/internal/models"
	"sync"
	"time"
)

type MockDB struct {
	mu             sync.RWMutex
	Users          map[uuid.UUID]*models.User
	UserPasswords  map[string]string // username -> hash
	Alphas         map[uuid.UUID]*models.Alpha
	BacktestRuns   map[uuid.UUID]*models.BacktestRun
	AuditLogs      []models.AuditLog
	Factors        map[uuid.UUID]*models.Factor
}

func NewMockDB() *MockDB {
	return &MockDB{
		Users:         make(map[uuid.UUID]*models.User),
		UserPasswords: make(map[string]string),
		Alphas:        make(map[uuid.UUID]*models.Alpha),
		BacktestRuns:  make(map[uuid.UUID]*models.BacktestRun),
		AuditLogs:     []models.AuditLog{},
		Factors:       make(map[uuid.UUID]*models.Factor),
	}
}

// User methods
func (m *MockDB) GetUserByUsername(ctx context.Context, username string) (*models.User, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, u := range m.Users {
		if u.Username == username {
			return u, m.UserPasswords[username], nil
		}
	}
	return nil, "", nil
}

func (m *MockDB) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if u, ok := m.Users[id]; ok {
		return u, nil
	}
	return nil, nil
}

func (m *MockDB) ListUsers(ctx context.Context) ([]models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := []models.User{}
	for _, u := range m.Users {
		res = append(res, *u)
	}
	return res, nil
}

func (m *MockDB) CreateUser(ctx context.Context, username, passwordHash string, role models.UserRole, fullName string) (*models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u := &models.User{
		ID:        uuid.New(),
		Username:  username,
		Role:      role,
		FullName:  fullName,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.Users[u.ID] = u
	m.UserPasswords[username] = passwordHash
	return u, nil
}

func (m *MockDB) CreateAuditLog(ctx context.Context, userID uuid.UUID, action models.AuditAction, entityType models.AuditEntityType, entityID uuid.UUID, details string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.AuditLogs = append(m.AuditLogs, models.AuditLog{
		ID:         len(m.AuditLogs) + 1,
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Details:    details,
		CreatedAt:  time.Now(),
	})
	return nil
}

func (m *MockDB) ListAuditLogs(ctx context.Context, limit, offset int) ([]models.AuditLog, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.AuditLogs, len(m.AuditLogs), nil
}

// Alpha methods
func (m *MockDB) CreateAlpha(ctx context.Context, authorID uuid.UUID, name, description, codeContent string) (*models.Alpha, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	a := &models.Alpha{
		ID:          uuid.New(),
		AuthorID:    authorID,
		Name:        name,
		Description: description,
		CodeContent: codeContent,
		Status:      models.AlphaStatusDraft,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.Alphas[a.ID] = a
	return a, nil
}

func (m *MockDB) GetAlphaByID(ctx context.Context, id uuid.UUID) (*models.Alpha, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if a, ok := m.Alphas[id]; ok {
		return a, nil
	}
	return nil, nil
}

func (m *MockDB) ListAlphasByAuthor(ctx context.Context, authorID uuid.UUID) ([]models.Alpha, error) {
	res := []models.Alpha{}
	for _, a := range m.Alphas {
		if a.AuthorID == authorID {
			res = append(res, *a)
		}
	}
	return res, nil
}

func (m *MockDB) ListSubmittedAlphas(ctx context.Context) ([]models.Alpha, error) {
	res := []models.Alpha{}
	for _, a := range m.Alphas {
		if a.Status == models.AlphaStatusSubmitted {
			res = append(res, *a)
		}
	}
	return res, nil
}

func (m *MockDB) SubmitAlpha(ctx context.Context, id uuid.UUID) (*models.Alpha, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if a, ok := m.Alphas[id]; ok {
		a.Status = models.AlphaStatusSubmitted
		a.UpdatedAt = time.Now()
		return a, nil
	}
	return nil, nil
}

// Backtest methods
func (m *MockDB) CreateBacktestRun(ctx context.Context, alphaID, executorID uuid.UUID, params map[string]interface{}) (*models.BacktestRun, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	b := &models.BacktestRun{
		ID:         uuid.New(),
		AlphaID:    alphaID,
		ExecutorID: executorID,
		Status:     models.JobStatusPending,
		Params:     params,
		CreatedAt:  time.Now(),
	}
	m.BacktestRuns[b.ID] = b
	return b, nil
}

func (m *MockDB) GetBacktestRunByID(ctx context.Context, id uuid.UUID) (*models.BacktestRun, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if b, ok := m.BacktestRuns[id]; ok {
		return b, nil
	}
	return nil, nil
}

func (m *MockDB) UpdateBacktestRunStatus(ctx context.Context, id uuid.UUID, status models.JobStatus, metrics map[string]interface{}, errorLog string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if b, ok := m.BacktestRuns[id]; ok {
		b.Status = status
		b.Metrics = metrics
		b.ErrorLog = errorLog
		now := time.Now()
		b.FinishedAt = &now
	}
	return nil
}

func (m *MockDB) UpdateUser(ctx context.Context, id uuid.UUID, username, fullName string, role models.UserRole) (*models.User, error) {
	if u, ok := m.Users[id]; ok {
		u.Username = username
		u.FullName = fullName
		u.Role = role
		u.UpdatedAt = time.Now()
		return u, nil
	}
	return nil, nil
}

func (m *MockDB) DeleteUser(ctx context.Context, id uuid.UUID) error {
	delete(m.Users, id)
	return nil
}

// Factor methods
func (m *MockDB) CreateFactor(ctx context.Context, name, description, dataPath, frequency string, createdBy uuid.UUID) (*models.Factor, error) {
	f := &models.Factor{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		DataPath:    dataPath,
		Frequency:   frequency,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.Factors[f.ID] = f
	return f, nil
}

func (m *MockDB) GetFactorByID(ctx context.Context, id uuid.UUID) (*models.Factor, error) {
	if f, ok := m.Factors[id]; ok {
		return f, nil
	}
	return nil, nil
}

func (m *MockDB) ListFactors(ctx context.Context) ([]models.Factor, error) {
	res := []models.Factor{}
	for _, f := range m.Factors {
		res = append(res, *f)
	}
	return res, nil
}

func (m *MockDB) UpdateFactor(ctx context.Context, id uuid.UUID, name, description, dataPath, frequency string) (*models.Factor, error) {
	if f, ok := m.Factors[id]; ok {
		f.Name = name
		f.Description = description
		f.DataPath = dataPath
		f.Frequency = frequency
		f.UpdatedAt = time.Now()
		return f, nil
	}
	return nil, nil
}

func (m *MockDB) DeleteFactor(ctx context.Context, id uuid.UUID) error {
	delete(m.Factors, id)
	return nil
}

func (m *MockDB) UpdateAlpha(ctx context.Context, id uuid.UUID, name, description, codeContent string) (*models.Alpha, error) {
	if a, ok := m.Alphas[id]; ok {
		a.Name = name
		a.Description = description
		a.CodeContent = codeContent
		a.UpdatedAt = time.Now()
		return a, nil
	}
	return nil, nil
}

func (m *MockDB) DeleteAlpha(ctx context.Context, id uuid.UUID) error {
	delete(m.Alphas, id)
	return nil
}

func (m *MockDB) ListBacktestRunsByExecutor(ctx context.Context, executorID uuid.UUID) ([]models.BacktestRun, error) {
	res := []models.BacktestRun{}
	for _, b := range m.BacktestRuns {
		if b.ExecutorID == executorID {
			res = append(res, *b)
		}
	}
	return res, nil
}

func (m *MockDB) ListBacktestRuns(ctx context.Context) ([]models.BacktestRun, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := []models.BacktestRun{}
	for _, b := range m.BacktestRuns {
		res = append(res, *b)
	}
	return res, nil
}

func (m *MockDB) DeleteJob(ctx context.Context, jobID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.BacktestRuns, jobID)
	return nil
}

func (m *MockDB) FindActiveBacktestRun(ctx context.Context, alphaID uuid.UUID, params map[string]interface{}) (*models.BacktestRun, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	paramsJSON, _ := json.Marshal(params)
	
	for _, b := range m.BacktestRuns {
		if b.AlphaID == alphaID && (b.Status == models.JobStatusPending || b.Status == models.JobStatusRunning) {
			bParamsJSON, _ := json.Marshal(b.Params)
			if string(paramsJSON) == string(bParamsJSON) {
				return b, nil
			}
		}
	}
	return nil, nil
}

// Model methods
func (m *MockDB) CreateModel(ctx context.Context, name, version string, dsID uuid.UUID, pklPath string, trainingParams map[string]interface{}) (*models.Model, error) {
	return nil, nil // Placeholder
}

func (m *MockDB) GetModelByID(ctx context.Context, id uuid.UUID) (*models.Model, error) {
	return nil, nil // Placeholder
}

func (m *MockDB) ListModels(ctx context.Context) ([]models.Model, error) {
	return nil, nil // Placeholder
}

func (m *MockDB) DeleteModel(ctx context.Context, id uuid.UUID) error {
	return nil // Placeholder
}

// Analytics methods
func (m *MockDB) GetPerformance(ctx context.Context) ([]models.PerformanceItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := []models.PerformanceItem{}
	for _, b := range m.BacktestRuns {
		if b.Status == models.JobStatusCompleted {
			alphaName := "Unknown"
			authorName := "Unknown"
			if a, ok := m.Alphas[b.AlphaID]; ok {
				alphaName = a.Name
				if u, ok := m.Users[a.AuthorID]; ok {
					authorName = u.Username
				}
			}

			item := models.PerformanceItem{
				AlphaID:    b.AlphaID,
				AlphaName:  alphaName,
				AuthorName: authorName,
				Status:     string(b.Status),
			}
			// Map metrics if available
			if mVal, ok := b.Metrics["total_pnl"].(float64); ok { item.TotalReturn = mVal }
			if mVal, ok := b.Metrics["sharpe_ratio"].(float64); ok { item.Sharpe = mVal }
			if mVal, ok := b.Metrics["win_rate"].(float64); ok { item.WinRate = mVal }
			if mVal, ok := b.Metrics["max_drawdown"].(float64); ok { item.MaxDrawdown = mVal }

			res = append(res, item)
		}
	}
	return res, nil
}

func (m *MockDB) GetCorrelationData(ctx context.Context) ([]models.CorrelationItemRaw, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := []models.CorrelationItemRaw{}
	for _, b := range m.BacktestRuns {
		if b.Status == models.JobStatusCompleted && b.Metrics["pnl_curve"] != nil {
			alphaName := "Unknown"
			if a, ok := m.Alphas[b.AlphaID]; ok {
				alphaName = a.Name
			}
			
			curve, _ := b.Metrics["pnl_curve"].([]map[string]interface{})
			res = append(res, models.CorrelationItemRaw{
				AlphaName: alphaName,
				PnLCurve:  curve,
			})
		}
	}
	return res, nil
}

// Ping for health checks
func (m *MockDB) Ping(ctx context.Context) error {
	return nil
}
