# Contributing to QuantAlpha

Welcome! This project follows a **Strict Verification First** engineering culture.

## 1. Development Workflow

1. **Plan**: Create or update the implementation plan before major changes.
2. **Implement**: Keep changes minimal, additive, and rollback-safe.
3. **Verify**: Every PR must pass the automated CI and manual smoke test.
4. **Document**: Update `RELEASE_BASELINE.md` if architecture or RBAC changes.

## 2. Standards

### Backend (Go)
- Run `gofmt` before committing.
- Ensure all handlers have robust error handling.
- Use structured logging.

### Frontend (Angular)
- Do not use build-time constants for environment data.
- Use `window.APP_CONFIG` for runtime configuration.
- Maintain type safety (no `any` without strong justification).

### Worker (Python)
- Use `py_compile` to verify syntax.
- Ensure graceful handling of Redis disconnects.

## 3. Pull Request Checklist

- [ ] Code is formatted (`gofmt`, `prettier`).
- [ ] No hardcoded localhost URLs in production paths.
- [ ] Smoke test passes locally with `docker compose`.
- [ ] Documentation reflects new changes.

## 4. Communication
Report all technical debt and regression risks immediately. Stability is prioritized over features.
