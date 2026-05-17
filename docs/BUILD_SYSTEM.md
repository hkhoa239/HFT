# QuantAlpha Build System Documentation

This document describes how the QuantAlpha stack is compiled, validated, and packaged.

## 1. Database Layer (Handwritten)

> [!WARNING]
> The file `backend/internal/database/queries.sql.go` is currently **handwritten**.

Although it follows naming conventions typical of tools like `sqlc`, there is currently no active code generation pipeline. All SQL changes must be performed manually in this file. 

**Future Plan**: Migration to `sqlc` is on the roadmap but is currently out of scope for the MVP stabilization.

## 2. Frontend Build (Runtime Config)

The frontend uses a two-stage approach to configuration:
1. **Build Time**: The app is compiled as a static SPA.
2. **Runtime**: A shell entrypoint (`entrypoint.sh`) uses `envsubst` to populate `assets/config.js` from environment variables before starting Nginx.

This pattern allows the same Docker image to be deployed to different environments without rebuilding.

## 3. Validation Pipeline

The CI system (`.github/workflows/ci.yml`) executes the following checks:
1. **Frontend**: Typecheck (`tsc`) + Formatting check (`prettier`).
2. **Backend**: Formatting check (`gofmt`) + Build (`go build`).
3. **Worker**: Syntax check (`py_compile`).
4. **Infra**: Compose config validation (`docker compose config`).

## 4. Local Automation

The `Makefile` (if available) or manual scripts provide local shortcuts:
- `make build`: Rebuilds all containers.
- `make smoke`: Runs the automated smoke test.
