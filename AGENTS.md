# AI Contributor Guide: DockSlim

This file is the primary guide for AI agents (such as OpenAI Codex) contributing to this repository. It explains what the project does, how the codebase is structured, which technologies to use, and how to work iteratively with small, testable changes.

Please read this carefully and follow these rules for **every** task.

---

## 1. Project Overview

**DockSlim** is a web service that analyzes Docker images and helps teams:

- Understand where space is being used inside their images.
- See per-layer and per-directory size breakdowns.
- Identify large or unnecessary files (e.g., `.git`, `node_modules`, `__pycache__`, build artifacts, logs).
- Get concrete optimization hints to reduce image size and speed up CI/CD.

High-level flow:

1. User registers and creates a **Project**.
2. User configures a container **Registry** (for now, a generic Docker Registry; later Docker Hub, GitLab/GitFlic, etc.).
3. User triggers an **Image Analysis** by specifying repository + tag.
4. A background **Analyzer** service pulls the image (or at least its metadata and layers), inspects layers, extracts file/directory sizes, and stores results in PostgreSQL.
5. The **Web UI** visualizes the analysis: total size, layer histogram, directory tree, top N largest files, and basic recommendations.

This repository should evolve into a small but production-minded SaaS-style application with:

- A Go backend API.
- A Go analyzer worker.
- A React + TypeScript frontend.
- PostgreSQL and Redis in the dev environment.
- Docker-based development setup (`docker-compose` initially).

---

## 2. Technologies and Tools

You MUST use the following tech stack unless a task explicitly states otherwise:

### Backend API

- Language: **Go** (latest stable, e.g. 1.22+)
- HTTP Router: a lightweight router (e.g. `chi`, `echo`, or `fiber`). Once chosen for the project, stick to it consistently.
- Database: **PostgreSQL**
- Cache / async coordination: **Redis** (initially can be used for queues, locks, or simple caching)
- Migrations: a Go-friendly migration tool (for example `golang-migrate` or similar)
- Config: environment variables with a small configuration package for parsing (e.g. `CONFIG_...` prefixes)
- Testing: Go `testing` package with table-driven tests

### Analyzer Worker

- Language: **Go**
- Purpose: pull Docker image metadata/layers, unpack tar layers, compute size per layer, directory, and top files.
- Interaction with backend:
  - For MVP, it can write analysis results directly into PostgreSQL using the same schema as the backend.
  - Later, we may use Redis or a message queue for coordination.

### Frontend

- Framework: **Vue 3**
- Language: **JavaScript** (with single-file components, `.vue`)
- Bundler/Dev Server: **Vite**
- Styling: **Tailwind CSS** (no heavy design needed, just clean and consistent)
- Routing: **Vue Router** (or similar) for basic pages.

### Dev / Ops

- Containerization: Docker
- Local orchestration: `docker-compose` for development (`Postgres`, `Redis`, backend API, analyzer, frontend).
- Optional later: Helm charts for Kubernetes deployment.

---

## 3. Repository Structure

The target structure (you can refine details, but keep the spirit):

- `/backend`
  - `cmd/api` – main HTTP API binary.
  - `internal/config` – configuration loading.
  - `internal/db` – DB connection, migrations, repositories.
  - `internal/http` – HTTP handlers, routing.
  - `internal/services` – business logic (use-cases).
  - `internal/registry` – clients for container registries.
  - `internal/analysis` – read/write access to analysis results.
  - `internal/auth` – authentication, user management (once implemented).

- `/analyzer`
  - `cmd/analyzer` – main worker binary.
  - `internal/analyzer` – image analysis logic.
  - `internal/registry` – registry client usage (can share code via a common module or reuse as needed).
  - `internal/db` – DB access reused or mirrored from backend.

- `/frontend`
  - `src/` – Vue 3 + JavaScript source code.
    - `main.js` – app entrypoint.
    - `App.vue` – root component.
    - `components/` – reusable UI components.
    - `views/` – page-level components (Login, Projects, Project details, Analysis details, etc.).
    - `api/` – API client wrappers for backend (if needed).
  - `index.html`, `vite.config.js` or `vite.config.mts`, Tailwind config, etc.

- `/deploy`
  - `docker-compose.yml` – dev stack: Postgres, Redis, backend, analyzer, frontend.
  - (Later) Kubernetes manifests, Helm charts, etc.

- `/docs`
  - Additional documentation and diagrams (optional).

- `AGENTS.md` – this file.
- `README.md` – general human-oriented documentation.

Do NOT create extra top-level folders unless necessary. Always update this section if the structure meaningfully changes.

---

## 4. Coding Conventions

### 4.1. Go (Backend and Analyzer)

- Use **Go modules** (one module for the repository is fine).
- Always run `go fmt` (or `gofmt`) on Go files; code must be formatted.
- Follow idiomatic Go patterns:
  - `if err != nil { ... }` error handling, no hidden panics unless absolutely needed.
  - Context usage (`context.Context`) for request/DB operations where appropriate.
- Organize Go packages by responsibility, not by technical layer names only. Keep packages small and cohesive.
- Avoid global state; pass dependencies via constructors (dependency injection).

### 4.2. Vue + JavaScript

- Use **Vue 3** with the Composition API (preferred) or Options API where appropriate.
- Use single-file components (`.vue`) with `<script>` + `<template>` (and optionally `<style>`).
- Keep components small; extract reusable UI into `/components`.
- Use Tailwind classes for layout and basic styling.
- Use JSDoc comments or clear prop definitions to keep component contracts understandable.

### 4.3. General

- Comments:
  - Add comments for non-obvious logic.
  - Do not overcomment trivial code.
- Naming:
  - Prefer clear, descriptive names over abbreviations.
  - For Go packages, use short, meaningful names (e.g., `analysis`, `registry`, `projects`).

---

## 5. Testing and Validation

Before considering a change “done”, you SHOULD:

1. For backend and analyzer:
   - Run: `go test ./...` inside the corresponding module (at least the packages you touched).
2. For frontend:
   - Ensure that the app builds:
     - `npm install` (or `pnpm`/`yarn` depending on lockfile)
     - `npm run build` (or equivalent)
   - Optionally run any configured tests or linters (e.g. `npm run lint` or `npm test`).

If a task modifies both backend and frontend, run the relevant tests/build commands for all modified parts.

If tests or builds fail, you must fix the issues as part of the same task instead of leaving the project in a broken state.

---

## 6. Dev Environment and Commands

The typical local development flow should be:

- Start services via Docker Compose (once defined):
  - `docker-compose up` (or `docker compose up`)
- Or run individual services locally:
  - Backend:
    - `cd backend`
    - `go run ./cmd/api`
  - Analyzer:
    - `cd analyzer`
    - `go run ./cmd/analyzer`
  - Frontend:
    - `cd frontend`
    - `npm install`
    - `npm run dev`

For every task you complete, please:

- Update docs/README if you introduce new commands or configuration options.
- Confirm that `docker-compose up` (or the documented dev commands) still work after your changes.

---

## 7. Workflow for AI Agents (VERY IMPORTANT)

We will use this repository with an AI agent iteratively. Follow these rules for **every** Code-mode task:

1. **Small, focused tasks**  
   - Keep each change set reasonably small (rough target: **500–1000 lines of code** or less, including tests).
   - Do not attempt to implement the entire product in one task.
   - Prefer vertical slices that deliver a testable feature (e.g., “health endpoint”, “basic user model and registration”, “list projects in UI”).

2. **Plan before coding**  
   - Start each task by:
     - summarizing the goal,
     - listing the files to change or create,
     - outlining the steps (e.g., “add DB migration”, “add handler”, “add frontend component”).

3. **Preserve working state**  
   - The repository MUST remain buildable and runnable after your changes.
   - Run tests/build commands and fix issues before presenting results.

4. **Do NOT rewrite the whole project**  
   - Do not reformat or rewrite large parts of the codebase that are unrelated to the requested change.
   - Do not delete or drastically refactor fundamental structure unless explicitly instructed.

5. **Output expectations for each task**  
   For each task, when you return the result, include:

   - A human-readable summary of the change.
   - A list of modified/created files.
   - Any new or changed commands (e.g. “run `go test ./backend/...`”, “run `npm run build`”).
   - Notes about any assumptions you made.
   - Any TODOs or limitations that remain.

6. **Respect this AGENTS.md**  
   - Do not modify this `AGENTS.md` file unless explicitly asked to.
   - If you believe the guide is missing important information, mention it in your task summary, but do not change the file yourself.

---

## 8. Feature Roadmap (High-Level)

This is a rough feature roadmap to guide implementation steps:

1. **Initial skeleton (MVP foundation)**  
   - Backend API skeleton with `/health` endpoint.
   - Analyzer worker skeleton that starts up and logs that it is running.
   - Frontend skeleton with a placeholder “DockSlim” page.
   - `docker-compose.yml` for dev with Postgres + Redis + services.
   - Basic README with instructions.

2. **Users & Projects**  
   - User registration and login (JWT or session).
   - Projects CRUD (for authenticated users).

3. **Registry connections**  
   - Model and API for registries.
   - For MVP: generic Docker Registry with URL + credentials.

4. **Image analysis pipeline (MVP version)**  
   - API endpoint to trigger an analysis (project + registry + repo + tag).
   - Analyzer implementation that:
     - pulls image (or layers),
     - extracts layer sizes,
     - computes simple directory-size tree,
     - stores results.
   - API to fetch analysis results.

5. **Frontend UI for analysis**  
   - List of analyses for a project.
   - Details page with:
     - total size,
     - layer histogram,
     - directory tree,
     - top files.

6. **Basic recommendations**  
   - Simple heuristics to flag common issues:
     - presence of `.git`, `node_modules`, `__pycache__`, `target`, large log files, etc.
     - base images that look unnecessarily heavy.

7. **Later enhancements**  
   - Image comparison (diff between two analyses).
   - CI/CD integration to gate pipelines by image size.
   - More advanced recommendation engine.
   - Multi-registry support (Docker Hub, GitLab/GitFlic, etc.).
   - Self-hosted + production deployment guides.

Use this roadmap to choose sensible next steps when implementing features.

---

## 9. Pull Requests and Branches (when applicable)

If you are asked to create a branch and a pull request:

- Branch names:
  - `feat/<short-description>` for features.
  - `fix/<short-description>` for bug fixes.
- Commit messages:
  - Start with a short verb phrase in present tense, e.g., `Add basic image analysis model`.
- PR title format:
  - `feat: short description` or `fix: short description`.
- PR description:
  - Brief summary of changes.
  - List of commands to run for testing.
  - Any known limitations or follow-ups.

---

By following this guide, you will act as a consistent, reliable contributor to the DockSlim project. Always prioritize small, testable increments that keep the repository in a clean, working state.
