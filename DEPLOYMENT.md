# QuantAlpha Deployment Guide

This guide provides instructions for deploying the QuantAlpha stack using Docker Compose.

## 1. Prerequisites
- Docker & Docker Compose (v2.0+)
- 4GB+ RAM available
- Internet access (for image pulls)

## 2. Local/Staging Startup

### Step 1: Environment Setup
Copy the example environment file:
```bash
cp .env.example .env
```

### Step 2: Build & Start
```bash
docker compose up --build -d
```

### Step 3: Verify
Run the automated smoke test:
```bash
python scripts/smoke-test.py
```

## 3. Configuration

### Backend Environment Variables
| Variable | Default | Description |
|----------|---------|-------------|
| `JWT_SECRET` | `quantalpha-secret` | Key used for JWT signing. |
| `GIN_MODE` | `release` | Set to `debug` for verbose logs. |

### Frontend Runtime Config
The frontend uses `envsubst` to inject these at startup:
- `API_URL`: Backend endpoint (e.g., `http://localhost:8080`)
- `APP_ENV`: `staging`, `production`, or `local`

## 4. Maintenance

### Database Backups
```bash
docker exec hft-postgres-1 pg_dump -U postgres quantalpha > backup.sql
```

### Log Inspection
```bash
docker compose logs -f backend
docker compose logs -f worker
```

## 5. Troubleshooting
- **Backend unhealthy**: Check if Postgres port 5432 is already bound on the host.
- **Frontend 404**: Ensure `nginx.conf` is correctly mapping the SPA routes.
- **Worker not consuming**: Verify Redis Stream connectivity in `docker compose logs worker`.
