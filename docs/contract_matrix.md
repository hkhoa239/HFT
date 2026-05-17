# API Contract Matrix

This matrix defines the stable fields and formats guaranteed by the Backend for Frontend (BFF) usage.

## Alpha Entity
| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `id` | UUID | Yes | Primary Key |
| `author_id` | UUID | Yes | Ownership field |
| `name` | String | Yes | Non-empty |
| `status` | Enum | Yes | Values: `draft`, `submitted` |
| `description`| String | No | Nullable |
| `created_at` | Date | Yes | ISO8601 |

## Backtest Run Entity
| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `id` | UUID | Yes | Job ID |
| `alpha_id` | UUID | Yes | Reference |
| `status` | Enum | Yes | Values: `pending`, `running`, `completed`, `failed` |
| `params` | JSON | Yes | Must contain `start`, `end`, `capital` |
| `metrics` | JSON | No | Populated after completion |

## Common Failure Modes
| Status Code | Meaning | Expected Body |
|-------------|---------|---------------|
| 400 | Validation Fail | `{ "success": false, "error": "..." }` |
| 401 | Unauthorized | `{ "success": false, "error": "unauthorized" }` |
| 409 | Conflict | `{ "success": false, "error": "already in progress" }` |
| 500 | Internal Err | `{ "success": false, "error": "..." }` |
