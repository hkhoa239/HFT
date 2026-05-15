# QuantAlpha Lab - Design Document (Phase 1)

## 1. Database DDL Script

### 1.1 Type Definitions

```sql
-- Drop existing types
DROP TYPE IF EXISTS user_role CASCADE;
DROP TYPE IF EXISTS job_status CASCADE;
DROP TYPE IF EXISTS alpha_status CASCADE;

-- New unified types
CREATE TYPE user_role AS ENUM ('admin', 'qr', 'pm', 'ds');
CREATE TYPE job_status AS ENUM ('pending', 'running', 'completed', 'failed');
CREATE TYPE alpha_status AS ENUM ('draft', 'submitted');

-- Audit types
CREATE TYPE entity_type AS ENUM ('user', 'alpha', 'backtest', 'model', 'factor');
CREATE TYPE audit_action AS ENUM ('create', 'update', 'delete', 'run', 'submit');
```

### 1.2 Table Definitions

```sql
-- Users Table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    role user_role DEFAULT 'qr',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Factors Table
CREATE TABLE factors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    data_path VARCHAR(255) NOT NULL,
    frequency VARCHAR(10) DEFAULT '1d',
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Alphas Table
CREATE TABLE alphas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    code_content TEXT NOT NULL,
    status alpha_status DEFAULT 'draft',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Backtest Runs Table
CREATE TABLE backtest_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alpha_id UUID REFERENCES alphas(id) ON DELETE CASCADE,
    executor_id UUID REFERENCES users(id),
    status job_status DEFAULT 'pending',
    params JSONB NOT NULL DEFAULT '{}',
    metrics JSONB DEFAULT '{}',
    error_log TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    finished_at TIMESTAMP
);

-- Models Table
CREATE TABLE models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    version VARCHAR(20) NOT NULL,
    ds_id UUID REFERENCES users(id),
    pkl_path VARCHAR(255) NOT NULL,
    training_metrics JSONB,
    training_params JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Audit Logs Table
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    action audit_action NOT NULL,
    entity_type entity_type,
    entity_id UUID,
    details TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 1.3 Update Triggers

```sql
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_factors_updated_at BEFORE UPDATE ON factors FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_alphas_updated_at BEFORE UPDATE ON alphas FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_models_updated_at BEFORE UPDATE ON models FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

---

## 2. Go Structs (API Models)

### 2.1 User Models

```go
package models

type UserRole string

const (
    RoleAdmin UserRole = "admin"
    RoleQR    UserRole = "qr"
    RolePM   UserRole = "pm"
    RoleDS   UserRole = "ds"
)

type User struct {
    ID          uuid.UUID `json:"id"`
    Username    string   `json:"username"`
    FullName    string   `json:"full_name,omitempty"`
    Role        UserRole `json:"role"`
    CreatedAt   string   `json:"created_at"`
    UpdatedAt   string   `json:"updated_at"`
}

type CreateUserRequest struct {
    Username    string   `json:"username" validate:"required,min=3,max=50"`
    Password   string   `json:"password" validate:"required,min=6"`
    FullName   string   `json:"full_name" validate:"max=100"`
    Role       UserRole `json:"role" validate:"required,oneof=admin qr pm ds"`
}

type UpdateUserRequest struct {
    Username  string   `json:"username" validate:"min=3,max=50"`
    FullName  string   `json:"full_name" validate:"max=100"`
    Role     UserRole `json:"role" validate:"oneof=admin qr pm ds"`
}
```

### 2.2 Auth Models

```go
package models

type LoginRequest struct {
    Username string `json:"username" validate:"required"`
    Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
    Token     string `json:"token"`
    ExpiresAt string `json:"expires_at"`
    User     *User  `json:"user"`
}
```

### 2.3 Factor Models

```go
package models

type Factor struct {
    ID          uuid.UUID `json:"id"`
    Name        string   `json:"name"`
    Description string   `json:"description,omitempty"`
    DataPath    string   `json:"data_path"`
    Frequency  string   `json:"frequency"`
    CreatedBy  uuid.UUID`json:"created_by"`
    CreatedAt  string   `json:"created_at"`
    UpdatedAt  string   `json:"updated_at"`
}

type CreateFactorRequest struct {
    Name        string `json:"name" validate:"required,min=1,max=100"`
    Description string `json:"description"`
    DataPath    string `json:"data_path" validate:"required"`
    Frequency  string `json:"frequency" validate:"oneof=1m 5m 15m 1h 1d"`
}

type UpdateFactorRequest struct {
    Name        string `json:"name" validate:"min=1,max=100"`
    Description string `json:"description"`
    DataPath    string `json:"data_path"`
    Frequency  string `json:"frequency" validate:"oneof=1m 5m 15m 1h 1d"`
}

type FactorPreviewResponse struct {
    Data []map[string]interface{} `json:"data"`
}
```

### 2.4 Alpha Models

```go
package models

type AlphaStatus string

const (
    AlphaStatusDraft     AlphaStatus = "draft"
    AlphaStatusSubmitted AlphaStatus = "submitted"
)

type Alpha struct {
    ID           uuid.UUID  `json:"id"`
    AuthorID    uuid.UUID `json:"author_id"`
    Name        string    `json:"name"`
    Description string    `json:"description,omitempty"`
    CodeContent string    `json:"code_content"`
    Status      AlphaStatus`json:"status"`
    CreatedAt   string    `json:"created_at"`
    UpdatedAt   string    `json:"updated_at"`
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
```

### 2.5 Backtest Models

```go
package models

type JobStatus string

const (
    JobStatusPending   JobStatus = "pending"
    JobStatusRunning JobStatus = "running"
    JobStatusCompleted JobStatus = "completed"
    JobStatusFailed  JobStatus = "failed"
)

type BacktestRun struct {
    ID          uuid.UUID              `json:"id"`
    AlphaID     uuid.UUID              `json:"alpha_id"`
    ExecutorID  uuid.UUID              `json:"executor_id"`
    Status      JobStatus              `json:"status"`
    Params      map[string]interface{}`json:"params"`
    Metrics     map[string]interface{}`json:"metrics,omitempty"`
    ErrorLog    string                `json:"error_log,omitempty"`
    CreatedAt   string                `json:"created_at"`
    FinishedAt string                `json:"finished_at,omitempty"`
}

type BacktestParams struct {
    Start   string  `json:"start" validate:"required"`
    End     string  `json:"end" validate:"required"`
    Capital float64 `json:"capital" validate:"required,gt=0"`
}

type RunBacktestRequest struct {
    AlphaID string           `json:"alpha_id" validate:"required,uuid"`
    Params *BacktestParams `json:"params" validate:"required,dive"`
}

type BacktestStatusResponse struct {
    Status    JobStatus              `json:"status"`
    Metrics   map[string]interface{}`json:"metrics,omitempty"`
    ErrorLog  string                `json:"error_log,omitempty"`
}
```

### 2.6 Model/ML Models

```go
package models

type Model struct {
    ID              uuid.UUID              `json:"id"`
    Name            string               `json:"name"`
    Version         string               `json:"version"`
    DSID            uuid.UUID            `json:"ds_id"`
    PklPath        string              `json:"pkl_path"`
    TrainingMetrics map[string]interface{}`json:"training_metrics,omitempty"`
    TrainingParams  map[string]interface{}`json:"training_params,omitempty"`
    CreatedAt       string             `json:"created_at"`
    UpdatedAt       string             `json:"updated_at"`
}

type TrainModelRequest struct {
    Name       string               `json:"name" validate:"required,min=1,max=100"`
    Version    string                `json:"version" validate:"required"`
    Params     map[string]interface{}`json:"params"`
}

type ModelPreviewResponse struct {
    Data []map[string]interface{} `json:"data"`
}
```

### 2.7 Analytics Models

```go
package models

type CorrelationResponse struct {
    Data []CorrelationItem `json:"data"`
}

type CorrelationItem struct {
    AlphaID1  uuid.UUID `json:"alpha_id_1"`
    AlphaID2  uuid.UUID `json:"alpha_id_2"`
    Value     float64  `json:"value"`
}

type PerformanceResponse struct {
    Data []PerformanceItem `json:"data"`
}

type PerformanceItem struct {
    AlphaID    uuid.UUID `json:"alpha_id"`
    TotalReturn float64  `json:"total_return"`
    Sharpe    float64  `json:"sharpe"`
    MaxDrawdown float64 `json:"max_drawdown"`
}
```

### 2.8 Audit Log Models

```go
package models

type AuditEntityType string

const (
    EntityUser    AuditEntityType = "user"
    EntityAlpha  AuditEntityType = "alpha"
    EntityBacktest AuditEntityType = "backtest"
    EntityModel  AuditEntityType = "model"
    EntityFactor AuditEntityType = "factor"
)

type AuditAction string

const (
    AuditCreate AuditAction = "create"
    AuditUpdate AuditAction = "update"
    AuditDelete AuditAction = "delete"
    AuditRun    AuditAction = "run"
    AuditSubmit AuditAction = "submit"
)

type AuditLog struct {
    ID         int            `json:"id"`
    UserID     uuid.UUID      `json:"user_id"`
    Action     AuditAction    `json:"action"`
    EntityType AuditEntityType `json:"entity_type"`
    EntityID  uuid.UUID      `json:"entity_id,omitempty"`
    Details    string         `json:"details,omitempty"`
    CreatedAt  string        `json:"created_at"`
}

type AuditLogResponse struct {
    Data []AuditLog `json:"data"`
    Total int       `json:"total"`
}
```

---

## 3. Redis Stream Payloads

### 3.1 Job Queue Payload Structure

```go
package redis

type JobPayload struct {
    JobID     string `json:"job_id"`
    TaskType  string `json:"task_type"` // "backtest", "train", "factor_compute"
    UserID    string `json:"user_id"`
    AlphaID   string `json:"alpha_id,omitempty"`
    Params    map[string]interface{} `json:"params"`
    CreatedAt string `json:"created_at"` // ISO8601
}
```

### 3.2 JSON Format Examples

**Backtest Job:**
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "task_type": "backtest",
  "user_id": "550e8400-e29b-41d4-a716-446655440001",
  "alpha_id": "550e8400-e29b-41d4-a716-446655440002",
  "params": {
    "start": "2023-01-01",
    "end": "2023-12-31",
    "capital": 100000.0
  },
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Train Model Job:**
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440003",
  "task_type": "train",
  "user_id": "550e8400-e29b-41d4-a716-446655440001",
  "params": {
    "model_name": "model_v1",
    "factor_names": ["factor_A", "factor_B"]
  },
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Factor Compute Job:**
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440004",
  "task_type": "factor_compute",
  "user_id": "550e8400-e29b-41d4-a716-446655440001",
  "params": {
    "factor_name": "momentum_20d",
    "frequency": "1d"
  },
  "created_at": "2024-01-15T10:30:00Z"
}
```

---

## 4. RBAC Matrix

### 4.1 Role Definitions

| Role | Description |
|------|------------|
| admin | Full system access, can access all endpoints |
| qr    | Quantitative Researcher - research and alpha development |
| ds    | Data Scientist - factor creation, model training |
| pm    | Portfolio Manager - monitoring and analytics |

### 4.2 Endpoint Access Matrix

| Endpoint | Method | Required Roles | Notes |
|----------|--------|-----------------|-------|
| **Public** |||
| /auth/login | POST | - | No auth required |
| **Private (Auth)** |||
| /models | GET | admin,qr,pm,ds | |
| /factors | GET | admin,qr,pm,ds | |
| /me/profile | GET | admin,qr,pm,ds | |
| **QR Endpoints** |||
| /alphas/me | GET | admin,qr | |
| /alphas/me/submitted | GET | admin,qr | |
| /alphas | POST | admin,qr | |
| /alphas/:id | PUT | admin,qr | Owner only |
| /alphas/:id/submit | POST | admin,qr | Owner only |
| /alphas/:id | DELETE | admin,qr | Owner only (soft delete) |
| /backtest/run | POST | admin,qr | |
| /backtest/me/status | GET | admin,qr | |
| /backtest/:job_id | GET | admin,qr | |
| /factors/:id/preview | GET | admin,qr | |
| **DS Endpoints** |||
| /factors/publish | POST | admin,ds | |
| /factors/:id | PUT | admin,ds | |
| /factors/:id | DELETE | admin,ds | Soft delete |
| /models/train | POST | admin,ds | |
| /models/:id | DELETE | admin,ds | Soft delete |
| **PM Endpoints** |||
| /alphas/submitted | GET | admin,pm | |
| /backtest/status | GET | admin,pm | |
| /analytics/correlation | GET | admin,pm | |
| /analytics/performance | GET | admin,pm | |
| /audit-logs | GET | admin,pm | |
| **Admin Endpoints** |||
| /admin/users | GET | admin | |
| /admin/users | POST | admin | |
| /admin/users/:id | PATCH | admin | |
| /admin/users/:id | DELETE | admin | Soft delete |
| /admin/jobs/:job_id | DELETE | admin | Cancel/Kill job |

### 4.3 Middleware Implementation Logic

```go
func RBACMiddleware(allowedRoles ...UserRole) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get user from JWT context
        userVal, exists := c.Get("user")
        if !exists {
            c.JSON(401, gin.H{"error": "unauthorized"})
            c.Abort()
            return
        }
        
        user := userVal.(*models.User)
        
        // Admin bypass - full access
        if user.Role == RoleAdmin {
            c.Next()
            return
        }
        
        // Check if user's role is allowed
        allowed := false
        for _, role := range allowedRoles {
            if user.Role == role {
                allowed = true
                break
            }
        }
        
        if !allowed {
            c.JSON(403, gin.H{"error": "forbidden: insufficient permissions"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### 4.4 Admin-Only Middleware

```go
func AdminOnlyMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        userVal, exists := c.Get("user")
        if !exists {
            c.JSON(401, gin.H{"error": "unauthorized"})
            c.Abort()
            return
        }
        
        user := userVal.(*models.User)
        
        if user.Role != RoleAdmin {
            c.JSON(403, gin.H{"error": "forbidden: admin only"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

---

## 5. API Response格式

### 5.1 Success Responses

```json
{
  "success": true,
  "data": { ... }
}
```

### 5.2 Error Responses

```json
{
  "success": false,
  "error": "具体的错误信息"
}
```

### 5.3 Validation Error Responses

```json
{
  "success": false,
  "error": "field_name: validation message"
}
```

---

## 6. Storage Conventions (MinIO)

### 6.1 Path Structure

| Type | Path Template | Example |
|------|---------------|---------|
| Factor Data | factors/{user_id}_{factor_name}.parquet | factors/550e8400_factor_momentum.parquet |
| Factor Preview | factors/{factor_name}_preview.json | factors/momentum_20d_preview.json |
| Model | models/{user_id}/{model_name}/v{version}/model.pkl | models/550e8400/lstm_model/v1/model.pkl |
| Temp Scripts | scripts/{job_id}.py | scripts/550e8400-e29b-41d4-a716-446655440000.py |

---

## 7. Implementation Checklist

- [ ] Project initialization (Go modules, Gin)
- [ ] Database schema (DDL scripts)
- [ ] sqlc setup and query generation
- [ ] JWT authentication
- [ ] RBAC middleware
- [ ] User CRUD endpoints
- [ ] Auth login endpoint
- [ ] Factor endpoints
- [ ] Alpha endpoints
- [ ] Backtest endpoints
- [ ] Model endpoints
- [ ] Analytics endpoints
- [ ] Audit log endpoints
- [ ] Redis producer
- [ ] Admin endpoints