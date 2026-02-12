# DockSlim

DockSlim - tool for inspecting Docker images contents and understanding where space is used. It surfaces layer and size insights so teams can prevent image growth regressions.

## Project Structure

- `backend/` – Go HTTP API (chi router).
- `analyzer/` – Go worker for image analysis jobs.
- `frontend/` – Vue 3 + Vite + Tailwind production UI (themes, modals, custom selects/datepicker).
- `deploy/` – Docker Compose development stack.

## Prerequisites

- Go 1.22+
- Node.js 20+
- npm 10+
- Docker and Docker Compose

## Configuration

Copy the example environment files and adjust as needed:

```bash
cp backend/.env.example backend/.env
cp analyzer/.env.example analyzer/.env
```

Environment variables:

- Backend: `BACKEND_HTTP_PORT` (default: 8080), `POSTGRES_DSN`, `AUTO_MIGRATE` (default: true), `MIGRATIONS_PATH` (default: `backend/migrations`), `CORS_ALLOWED_ORIGINS` (comma-separated, default: `http://localhost:5173,http://127.0.0.1:5173`), `COOKIE_SECURE` (default: false), `COOKIE_SAMESITE` (`lax`, `strict`, or `none`, default: `lax`, requires `COOKIE_SECURE=true` when set to `none`), `COOKIE_DOMAIN` (optional), `COOKIE_PATH` (default: `/`), `INTERNAL_SUBSCRIPTION_TOKEN` (required for internal subscription update flow), `DOCKSLIM_BOOTSTRAP_ADMIN_EMAIL`, `DOCKSLIM_BOOTSTRAP_ADMIN_USERNAME`, `DOCKSLIM_BOOTSTRAP_ADMIN_PASSWORD`, `DOCKSLIM_BOOTSTRAP_ADMIN_PASSWORD_HASH`
- Analyzer: `ANALYZER_POSTGRES_DSN`
- Frontend: `VITE_API_BASE_URL` (optional override for API base URL), `VITE_API_PROXY_TARGET` (Vite dev proxy target, default: `http://localhost:8080`)

## Running Locally

### Backend API

Migrations run automatically on startup when `AUTO_MIGRATE` is true (default). To start the API locally:

```bash
cd backend
POSTGRES_DSN="postgres://dockslim:dockslim@localhost:5432/dockslim?sslmode=disable" \
BACKEND_HTTP_PORT=8080 \
AUTO_MIGRATE=true \
MIGRATIONS_PATH=migrations \
go run ./cmd/api
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Open `http://localhost:5173` in the browser.

Health check:

```bash
curl http://localhost:8080/health
```

Authentication flows:

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password123"}'

# Login (used for UI access; API calls below rely on personal API tokens)
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password123"}'

# Fetch the current user with an API token (preferred for API use)
API_TOKEN="ds_api_..." # create via /account/settings in the UI
curl -H "Authorization: Bearer ${API_TOKEN}" http://localhost:8080/api/v1/account/me
```

### Admin bootstrap (first run)

There are **no default admin credentials** and no migration-based admin seeding.

To bootstrap the first admin on startup, set:

```bash
DOCKSLIM_BOOTSTRAP_ADMIN_EMAIL=admin@example.com
DOCKSLIM_BOOTSTRAP_ADMIN_USERNAME=platform-admin
DOCKSLIM_BOOTSTRAP_ADMIN_PASSWORD='change-me-strong'
# or provide bcrypt hash instead of plaintext:
# DOCKSLIM_BOOTSTRAP_ADMIN_PASSWORD_HASH='$2a$10$...'
```

Rules:
- bootstrap runs only when the database has no admin users (`users.is_admin = true`);
- if the email already exists, that user is promoted to admin;
- if an admin already exists, bootstrap does nothing (idempotent).

Grant admin later (SQL):

```sql
UPDATE users
SET is_admin = TRUE, updated_at = NOW()
WHERE email = 'user@example.com';
```

Check admin users:

```sql
SELECT id, email, login, is_admin
FROM users
ORDER BY created_at ASC;
```

Projects API:

```bash
# Create project
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -d '{"name": "My Project"}'

# List projects
curl -H "Authorization: Bearer ${API_TOKEN}" http://localhost:8080/api/v1/projects

# Get project by ID
PROJECT_ID="your-project-id"
curl -H "Authorization: Bearer ${API_TOKEN}" http://localhost:8080/api/v1/projects/${PROJECT_ID}

# Update project name (owner only)
curl -X PATCH http://localhost:8080/api/v1/projects/${PROJECT_ID} \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -d '{"name": "Renamed Project"}'

# Delete project (owner only)
curl -X DELETE http://localhost:8080/api/v1/projects/${PROJECT_ID} \
  -H "Authorization: Bearer ${API_TOKEN}"
```

### Personal API tokens (user scoped)

Personal tokens let you call the API without CSRF headers.

**Create / revoke in UI**
1. Log in to the web app.
2. Open **Account → Account settings** (`/account/settings`).
3. In *Personal API tokens*, choose a name and create. Copy the token immediately (it is shown only once).
4. Revoke from the same table when you’re done.

**Use in API calls**

```
API_TOKEN="ds_api_..." # value shown once in the UI

# Example: list your projects without CSRF
curl -H "Authorization: Bearer ${API_TOKEN}" \
  http://localhost:8080/api/v1/projects

# Example: update profile
curl -X PATCH http://localhost:8080/api/v1/account/me \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"login":"new-handle","email":"me@example.com"}'
```

Notes:
- API tokens are user-scoped (not project/CI). They bypass CSRF.
- Tokens are rejected after revocation or expiry; use HTTPS in production.

### Subscriptions & billing

DockSlim now has built-in plans with server-side feature gating:

- `free`
- `pro`
- `team`

Subscription API:

```bash
# Current user subscription + resolved features/limits
curl -H "Authorization: Bearer ${API_TOKEN}" \
  http://localhost:8080/api/v1/account/subscription
```

Response shape:
- `plan` (`id`, `name`, `status`, `valid_until`, `is_admin`)
- `features` (resolved plan features)
- `limits` (derived limits like `history_days_limit`, `ci_comments`)

Internal update API (for future billing/webhook integration):

```bash
curl -X PUT http://localhost:8080/api/v1/internal/subscriptions \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -H "X-DockSlim-Internal-Token: ${INTERNAL_SUBSCRIPTION_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"<uuid>","plan_id":"pro","status":"active","valid_until":"2026-12-31T23:59:59Z"}'
```

Notes:
- `PUT /api/v1/internal/subscriptions` is restricted to admin users and requires `X-DockSlim-Internal-Token`.
- Regular frontend users cannot override plan directly.
- Billing UI page: `/account/billing`.

Feature visibility policy:
- `free`: keeps basic warnings (`insights.warnings`) and strips recommendations from gated outputs (`GET analysis`, JSON/PDF exports, CI compare report response).
- `free`: only non-advanced insight content is returned; advanced-prefixed/root advanced fields are removed and `insights` is reduced to sanitized warnings to avoid accidental leakage.
- `pro` / `team`: include full recommendations and advanced insights according to plan features.
- Export gates still apply by plan (`export_json`, `export_pdf`, etc.); when an export is allowed but `advanced_insights` is disabled, payload is sanitized using the same rules.

Registries API (project owners)

```bash
# List registries
curl -H "Authorization: Bearer ${API_TOKEN}" \
  http://localhost:8080/api/v1/projects/${PROJECT_ID}/registries

# Create registry
curl -X POST http://localhost:8080/api/v1/projects/${PROJECT_ID}/registries \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -d '{"name":"prod","type":"generic","registry_url":"https://registry.example.com","username":"ci","password":"token"}'

# Update registry
curl -X PATCH http://localhost:8080/api/v1/projects/${PROJECT_ID}/registries/${REGISTRY_ID} \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -d '{"name":"prod-eu","registry_url":"https://eu.registry.example.com","username":"ci","token":"new-token"}'

# Delete registry
curl -X DELETE http://localhost:8080/api/v1/projects/${PROJECT_ID}/registries/${REGISTRY_ID} \
  -H "Authorization: Bearer ${API_TOKEN}"
```

Budgets API (project owners)

```bash
# Get budgets (default + overrides)
curl -H "Authorization: Bearer ${API_TOKEN}" \
  http://localhost:8080/api/v1/projects/${PROJECT_ID}/budgets

# Upsert default budget
curl -X PUT http://localhost:8080/api/v1/projects/${PROJECT_ID}/budgets/default \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -d '{"warn_delta_mb":50,"fail_delta_mb":150,"hard_limit_mb":2048}'

# Create per-image override
curl -X POST http://localhost:8080/api/v1/projects/${PROJECT_ID}/budgets/overrides \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -d '{"image":"company/app","warn_delta_mb":30,"fail_delta_mb":120}'

# Update override
curl -X PATCH http://localhost:8080/api/v1/projects/${PROJECT_ID}/budgets/overrides/${BUDGET_ID} \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -d '{"fail_delta_mb":140}'

# Delete override
curl -X DELETE http://localhost:8080/api/v1/projects/${PROJECT_ID}/budgets/overrides/${BUDGET_ID} \
  -H "Authorization: Bearer ${API_TOKEN}"
```

Analysis comparison:

```bash
FROM_ANALYSIS_ID="analysis-id-a"
TO_ANALYSIS_ID="analysis-id-b"
curl -H "Authorization: Bearer ${API_TOKEN}" \
  "http://localhost:8080/api/v1/projects/${PROJECT_ID}/analyses/compare?from=${FROM_ANALYSIS_ID}&to=${TO_ANALYSIS_ID}"
```

In the frontend, open a completed analysis or the project analyses list and use the Compare action to see the size and layer diff between two completed analyses of the same image.

History, trends, and baselines:

- **History page** (`/projects/:id/history`): browse analyses across all statuses (queued, running, failed, completed) with image/branch/date filters.
- **Trends page** (`/projects/:id/trends`): visualize size and layer growth over time for a selected metric.
- **Baseline compare** (`/projects/:id/analyses/:analysisId`): compare an analysis against the latest completed analysis on the baseline branch (default: `main`). The current analysis is never used as its own baseline.

History API:

```bash
# List history (all statuses by default)
curl -H "Authorization: Bearer ${API_TOKEN}" \
  "http://localhost:8080/api/v1/projects/${PROJECT_ID}/history?image=repo/app&git_ref=main&status=all&from=2026-02-01&to=2026-02-07&limit=100"
```

Trends API:

```bash
# Fetch trends for a metric (total_size_bytes | layer_count | largest_layer_bytes)
curl -H "Authorization: Bearer ${API_TOKEN}" \
  "http://localhost:8080/api/v1/projects/${PROJECT_ID}/trends?metric=total_size_bytes&image=repo/app&git_ref=main&from=2026-02-01&to=2026-02-07&limit=500"
```

Baseline compare API:

```bash
ANALYSIS_ID="analysis-id"
curl -H "Authorization: Bearer ${API_TOKEN}" \
  "http://localhost:8080/api/v1/analyses/${ANALYSIS_ID}/baseline-compare"
```

Notes:
- `status` supports `all`, `completed`, `failed`, `running`, `queued`.
- `from`/`to` accept `YYYY-MM-DD` (inclusive) or RFC3339 timestamps.
- Baseline selection uses `baseline.mode` + `baseline.ref_branch` (currently `main_latest` on `main`) and excludes the current analysis. If no baseline exists, the API returns `404` with `no baseline analysis found`.
- Feature gating is enforced server-side:
  - `baseline-compare` requires `baseline_sla`.
  - `largest_layer_bytes` trends require `advanced_trends`.
  - free plan history is capped by `history_days_limit` (30 days by default).

Export endpoints:

```bash
# JSON export (requires export_json)
curl -H "Authorization: Bearer ${API_TOKEN}" \
  http://localhost:8080/api/v1/projects/${PROJECT_ID}/analyses/${ANALYSIS_ID}/export/json

# PDF export (requires export_pdf)
curl -H "Authorization: Bearer ${API_TOKEN}" \
  http://localhost:8080/api/v1/projects/${PROJECT_ID}/analyses/${ANALYSIS_ID}/export/pdf
```

### CI tokens & automation

Project owners can issue project-scoped CI tokens to let pipelines run analyses, generate compare reports, and post PR/MR comments without a user session.

```bash
# Use the token (Authorization: Bearer ds_ci_<...>)
CI_TOKEN="ds_ci_..."

# Trigger analysis & get report-friendly payload
curl -X POST http://localhost:8080/api/v1/ci/reports/image \
  -H "Authorization: Bearer ${CI_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"project_id":"'"${PROJECT_ID}"'","registry_id":"'"${REGISTRY_ID}"'","image":"repo/app","tag":"main"}'

# Compare two analyses and get markdown/json report + budget verdict
curl -X POST http://localhost:8080/api/v1/ci/reports/compare \
  -H "Authorization: Bearer ${CI_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"project_id":"'"${PROJECT_ID}"'","from_analysis_id":"'"${FROM_ANALYSIS_ID}"'","to_analysis_id":"'"${TO_ANALYSIS_ID}"'","include_markdown":true,"include_json":true}'

# Post PR/MR comment (no SCM token is stored)
curl -X POST http://localhost:8080/api/v1/ci/comments \
  -H "Authorization: Bearer ${CI_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"project_id":"'"${PROJECT_ID}"'","provider":"github","repo":"org/repo","pr_number":123,"scm_token":"ghp_...","body_markdown":"Report body"}'
```

### CI tokens API (owners)

```bash
# List CI tokens (metadata only)
curl -H "Authorization: Bearer ${API_TOKEN}" \
  http://localhost:8080/api/v1/projects/${PROJECT_ID}/ci-tokens

# Create CI token (plaintext token returned once)
curl -X POST http://localhost:8080/api/v1/projects/${PROJECT_ID}/ci-tokens \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${API_TOKEN}" \
  -d '{"name":"ci-runner","expires_at":"2026-03-01T00:00:00Z"}'

# Revoke CI token
curl -X POST http://localhost:8080/api/v1/projects/${PROJECT_ID}/ci-tokens/${TOKEN_ID}/revoke \
  -H "Authorization: Bearer ${API_TOKEN}"
```

Registry lookup for CI:
- `registry_id` has priority.
- or `registry_name`
- or `registry_host` (matches registries in the same project by hostname).
Missing/unknown returns 400/404; ambiguous host returns 409.

### Budgets / Limits

Budgets let you set size guardrails per project (default) and per image override. Verdicts appear on the Compare view (OK/WARN/FAIL) and API.

Endpoints (all project owners only for mutations):
- `GET /api/v1/projects/{projectId}/budgets` – fetch default + overrides.
- `PUT /api/v1/projects/{projectId}/budgets/default` – upsert default thresholds `{warn_delta_mb?, fail_delta_mb?, hard_limit_mb?}`.
- `POST /api/v1/projects/{projectId}/budgets/overrides` – create override `{image, warn_delta_mb?, fail_delta_mb?, hard_limit_mb?}`.
- `PATCH /api/v1/projects/{projectId}/budgets/overrides/{budgetId}` – update override.
- `DELETE /api/v1/projects/{projectId}/budgets/overrides/{budgetId}` – delete override.

Rules:
- Thresholds are in MB on the API; stored as bytes.
- Warnings trigger on regression > `warn_delta_mb`; FAIL on `fail_delta_mb` or exceeding `hard_limit_mb`.
- Per-image override wins over project default.
- Duplicate project name, registry name, or budget override returns HTTP 409 with a descriptive error.

Frontend:
- Project Settings → “Budgets” section: edit default thresholds and per-image overrides (owners only, others read-only).
- Compare page now shows “Budget verdict” badge and reasons alongside the Impact card.

### Analyzer Worker

```bash
cd analyzer
ANALYZER_POSTGRES_DSN="postgres://dockslim:dockslim@localhost:5432/dockslim?sslmode=disable" \
ANALYZER_REDIS_ADDR="localhost:6379" \
go run ./cmd/analyzer
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Then open http://localhost:5173 to view the DockSlim UI.

The Vite dev server proxies `/api` and `/health` to the backend (configured via `VITE_API_PROXY_TARGET`), so the frontend can reach the API without extra CORS setup in development.

## Docker Compose Dev Stack

You can run the full stack (Postgres, Redis, backend, analyzer, frontend) with Docker Compose. Migrations are applied automatically on backend startup using a PostgreSQL advisory lock, so no manual commands are required:

```bash
cd deploy
docker-compose up
```

Services are available at:

- Backend API: http://localhost:8080
- Frontend (Vite dev server): http://localhost:5173
- Postgres: localhost:5432 (user/password: dockslim)
- Redis: localhost:6379

Use `curl http://localhost:8080/health` to verify the backend is running.

### Tests

Run backend tests from the repository root:

```bash
go test ./...
```
