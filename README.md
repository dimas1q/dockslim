# DockSlim

DockSlim is a tool for inspecting Docker image contents and understanding where space is used. The project aims to surface layer and directory size insights to help teams trim their container images.

## Project Structure

- `backend/` – Go HTTP API (chi router).
- `analyzer/` – Go worker for image analysis jobs.
- `frontend/` – Vue 3 + Vite + Tailwind placeholder UI.
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

- Backend: `BACKEND_HTTP_PORT` (default: 8080), `POSTGRES_DSN`, `AUTO_MIGRATE` (default: true), `MIGRATIONS_PATH` (default: `backend/migrations`), `CORS_ALLOWED_ORIGINS` (comma-separated, default: `http://localhost:5173,http://127.0.0.1:5173`)
- Analyzer: `ANALYZER_POSTGRES_DSN`, `ANALYZER_REDIS_ADDR`
- Frontend: `VITE_API_BASE_URL` (default: `http://localhost:8080`)

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

# Login and capture the access token
ACCESS_TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password123"}' | jq -r '.access_token')

# Fetch the current user
curl -H "Authorization: Bearer ${ACCESS_TOKEN}" http://localhost:8080/api/v1/me
```

Projects API:

```bash
# Create project
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -d '{"name": "My Project"}'

# List projects
curl -H "Authorization: Bearer ${ACCESS_TOKEN}" http://localhost:8080/api/v1/projects

# Get project by ID
PROJECT_ID="your-project-id"
curl -H "Authorization: Bearer ${ACCESS_TOKEN}" http://localhost:8080/api/v1/projects/${PROJECT_ID}

# Update project name (owner only)
curl -X PATCH http://localhost:8080/api/v1/projects/${PROJECT_ID} \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -d '{"name": "Renamed Project"}'

# Delete project (owner only)
curl -X DELETE http://localhost:8080/api/v1/projects/${PROJECT_ID} \
  -H "Authorization: Bearer ${ACCESS_TOKEN}"
```

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

The backend allows CORS requests from `http://localhost:5173` by default so the frontend can call the API during development.

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
