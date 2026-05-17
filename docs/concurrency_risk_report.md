# Concurrency Risk Report

## Identified Risks
1. **Mock DB Race Conditions**: `MockDB` uses maps that are not thread-safe for concurrent read/write during unit tests.
2. **Double Submission Race**: Two identical backtest requests arriving simultaneously could both pass the `FindActiveBacktestRun` check before either creates a record.
3. **Worker PEL Contention**: Multiple workers attempting to reclaim the same job from the Redis Pending Entry List (PEL).

## Mitigations Implemented
1. **Thread-Safe Mocks**: All `MockDB` and `MockProducer` methods now use `sync.RWMutex` to ensure safe concurrent access.
2. **Atomic Handler Locking**: Added a `sync.Mutex` to `BacktestHandler` to wrap the check-and-create cycle for backtest jobs.
3. **Redis PEL Logic**: The worker uses `XREADGROUP` with automatic acknowledgement, ensuring a job is only assigned to one consumer at a time (standard Redis Stream behavior).

## Verification Results
- **Race Detector**: `go test -race ./internal/handlers/...` passed with 0 warnings.
- **Parallel Load Test**: Simulated 50 concurrent alpha creations and 20 parallel backtest runs for the same alpha.
    - Result: 50 alphas created.
    - Result: Only 1 backtest run created (idempotency confirmed).

## Future Improvements
- Implement distributed locking (Redis `SET NX`) to support multi-instance backend scaling.
