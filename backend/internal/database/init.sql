-- QuantAlpha HFT System - Initial Schema & Seed Data

-- 1. Type Definitions
DROP TYPE IF EXISTS user_role CASCADE;
DROP TYPE IF EXISTS job_status CASCADE;
DROP TYPE IF EXISTS alpha_status CASCADE;
DROP TYPE IF EXISTS entity_type CASCADE;
DROP TYPE IF EXISTS audit_action CASCADE;

CREATE TYPE user_role AS ENUM ('admin', 'qr', 'pm', 'ds', 'viewer');
CREATE TYPE job_status AS ENUM ('pending', 'running', 'completed', 'failed');
CREATE TYPE alpha_status AS ENUM ('draft', 'submitted');
CREATE TYPE entity_type AS ENUM ('user', 'alpha', 'backtest', 'model', 'factor');
CREATE TYPE audit_action AS ENUM ('create', 'update', 'delete', 'run', 'submit');

-- 2. Table Definitions

-- Users Table
CREATE TABLE IF NOT EXISTS users (
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
CREATE TABLE IF NOT EXISTS factors (
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
CREATE TABLE IF NOT EXISTS alphas (
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
CREATE TABLE IF NOT EXISTS backtest_runs (
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
CREATE TABLE IF NOT EXISTS models (
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
CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    action audit_action NOT NULL,
    entity_type entity_type,
    entity_id UUID,
    details TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. Triggers for updated_at
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

-- Seed Data
-- Default password for all seed users: password123
-- BCrypt hash: $2a$10$do0AgtKkvmPVk09CFxu.EuNcyrqJiOfa1IL5PqMrcdyz4TNzmF.mO

-- 1. admin
INSERT INTO users (username, password_hash, full_name, role)
VALUES (
    'admin', 
    '$2a$10$do0AgtKkvmPVk09CFxu.EuNcyrqJiOfa1IL5PqMrcdyz4TNzmF.mO', 
    'System Administrator', 
    'admin'
) ON CONFLICT DO NOTHING;

-- 2. quant (QR)
INSERT INTO users (username, password_hash, full_name, role)
VALUES (
    'quant', 
    '$2a$10$do0AgtKkvmPVk09CFxu.EuNcyrqJiOfa1IL5PqMrcdyz4TNzmF.mO', 
    'Lead Quant Researcher', 
    'qr'
) ON CONFLICT DO NOTHING;

-- 3. pm (Portfolio Manager)
INSERT INTO users (username, password_hash, full_name, role)
VALUES (
    'pm', 
    '$2a$10$do0AgtKkvmPVk09CFxu.EuNcyrqJiOfa1IL5PqMrcdyz4TNzmF.mO', 
    'Senior Portfolio Manager', 
    'pm'
) ON CONFLICT DO NOTHING;

-- 4. viewer
INSERT INTO users (username, password_hash, full_name, role)
VALUES (
    'viewer', 
    '$2a$10$do0AgtKkvmPVk09CFxu.EuNcyrqJiOfa1IL5PqMrcdyz4TNzmF.mO', 
    'Read-only Auditor', 
    'viewer'
) ON CONFLICT DO NOTHING;
