# QuantAlpha MVP - Release Engineering Baseline

This document defines the formal engineering baseline for the QuantAlpha MVP.

## 1. System Architecture

The system is designed as a modular, containerized HFT stack:

- **Frontend (Angular)**: SPA served by Nginx with runtime environment configuration.
- **Backend (Go)**: High-performance Gin API managing state, auth, and job orchestration.
- **Worker (Python)**: Event-driven backtest engine consuming tasks from Redis.
- **Redis**: Low-latency message broker using Streams for job distribution.
- **Postgres**: Canonical persistence layer with RBAC and audit logging.

## 2. Security & RBAC

### Role-Based Access Control (RBAC)
| Role | Description | Access |
|------|-------------|--------|
| `admin` | System Administrator | Full access to users, jobs, and all entities. |
| `qr` | Quant Researcher | Create alphas, run backtests, view own research. |
| `pm` | Portfolio Manager | View all submitted alphas, analytics, and audit logs. |
| `viewer` | Auditor | Read-only access to PM dashboard and analytics. |

### Seed Credentials (Baseline)
> [!IMPORTANT]
> Change all passwords in production environments.

| Username | Password | Role |
|----------|----------|------|
| `admin` | `password123` | `admin` |
| `quant` | `password123` | `qr` |
| `pm` | `password123` | `pm` |
| `viewer` | `password123` | `viewer` |

## 3. Operational Workflow

### Startup Sequence
1. Infrastructure (`Postgres`, `Redis`)
2. `Backend` (waits for PG/Redis health)
3. `Worker` (waits for PG/Redis health)
4. `Frontend` (waits for Backend health, performs `envsubst` for config)

### Smoke Testing
Automated verification is performed via `scripts/smoke-test.py`:
- Checks `/ready` endpoint.
- Validates `admin` login & JWT issuance.
- Verifies authenticated API access to `/alphas/me`.

## 4. Release Constraints
- **Business Logic**: Frozen for QR/PM modules.
- **Schema**: Managed via `init.sql`. No auto-migrations in MVP.
- **Config**: Environment-driven at runtime. No build-time constants.
