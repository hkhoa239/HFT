# LOCAL RUNBOOK: QuantAlpha Lab MVP

## 1. Infrastructure (Order Matters)
- **Redis**: Start local Redis service or use Docker:
  ```bash
  docker run -d -p 6379:6379 redis:alpine
  ```
- **Postgres**: Ensure the `hft` database is running and accessible.
  - Connection: `postgresql://hft:hft@localhost:5432/hft`

## 2. Services Startup
- **Backend (Go)**:
  ```bash
  cd backend
  go run cmd/server/main.go
  ```
  - Port: 8080
- **Worker (Python)**:
  ```bash
  cd worker
  python main.py
  ```
  - Dependencies: `pip install -r requirements.txt`
- **Frontend (Angular)**:
  ```bash
  cd frontend
  npm run dev
  ```
  - Port: 4200

## 3. Health Checks
- **Auth**: Login at `localhost:4200` with QR credentials.
- **Redis**: Run `redis-cli XINFO STREAM job_queue` to verify stream creation.
- **Worker**: Verify console output: "Consumer group worker_group created...".

## 4. Smoke Test
1. Navigate to **QR Workspace**.
2. Edit Alpha code or use default.
3. Click **RUN BACKTEST**.
4. Verify progress: `Initializing -> Job Queued -> Processing -> Completed`.
5. Verify PnL Chart rendering.
