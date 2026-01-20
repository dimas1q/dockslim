## 1. Product Vision (IMPORTANT)

**DockSlim** is a SaaS-style product for analyzing container images and helping
teams **understand, optimize, and control image size growth**.

The product answers three core questions:

1. Where is image size coming from?
2. Why did the image grow or change?
3. What should I do to fix or prevent it?

DockSlim is NOT just a technical analyzer.
It must always prioritize **actionable insights** and **clear UX**.

---

## 2. High-Level Architecture

DockSlim consists of four main components:

1. **Backend API (Go)**
   - Authentication, projects, registries, analyses
   - Enqueues analysis jobs
   - Serves analysis results and comparisons

2. **Analyzer Worker (Go)**
   - Background job processor
   - Fetches image manifests from registries
   - Computes layer breakdown, size metrics, insights, recommendations
   - Writes results back to PostgreSQL

3. **Frontend (Vue 3 + Vite + Tailwind)**
   - Project & analysis management
   - Rich visualization of results
   - Comparison and recommendations UI

4. **PostgreSQL**
   - Primary data store
   - Used as a job queue (analysis_jobs)
   - No external message broker is used at this stage

---

## 3. Current Analysis Pipeline (FACTUAL)

1. User creates an **Image Analysis** (image + tag + registry).
2. Backend:
   - creates image_analyses row (status = queued)
   - inserts analysis_jobs row
3. Analyzer worker:
   - picks jobs using `SELECT … FOR UPDATE SKIP LOCKED`
   - fetches image manifest (Docker v2 / OCI)
   - extracts layer sizes
   - computes total size, insights, recommendations
   - updates analysis status (completed / failed)
4. Frontend:
   - polls while queued/running
   - renders structured results when completed

---

## 4. Technologies (LOCKED)

### Backend & Analyzer
- Language: **Go (>=1.22)**
- Database: **PostgreSQL**
- Migrations: `golang-migrate`
- Auth:
  - Cookie-based session auth
  - CSRF protection
  - Key rotation-ready design
- Job Queue:
  - PostgreSQL (analysis_jobs)
  - No Redis / RabbitMQ for jobs

### Frontend
- Framework: **Vue 3**
- Language: **JavaScript**
- Build tool: **Vite**
- Styling: **Tailwind CSS**
- UX goal: modern, clean, SaaS-grade UI

### Dev Environment
- Docker + docker-compose
- Services: postgres, backend, analyzer, frontend

---

## 5. Repository Structure (ACTUAL)

```

/backend
/cmd/api
/cmd/migrate
/internal
/analysis
/auth
/config
/db
/httpapi
/projects
/registries

/analyzer
/cmd/analyzer
/internal
/analysis
/registry
/db

/frontend
/src
/components
/views
/api

/deploy
docker-compose.yml

AGENTS.md
README.md

````

Do not create new top-level directories unless explicitly required.

---

## 6. Analysis Result Contract (IMPORTANT)

Analyzer writes results to `image_analyses.result_json`.

Current canonical structure:

```json
{
  "image": "company/app",
  "tag": "1.2.3",
  "media_type": "docker",
  "layers": [
    {
      "digest": "sha256:…",
      "size_bytes": 123456,
      "media_type": "application/vnd.docker.image.rootfs.diff.tar.gzip"
    }
  ],
  "total_size_bytes": 12345678,
  "insights": {
    "layer_count": 12,
    "largest_layers": [ ... ],
    "warnings": [ ... ]
  },
  "recommendations": [ ... ]
}
````

Frontend relies on this structure.
Do NOT break it without coordination.

---

## 7. UX & Product Rules (VERY IMPORTANT)

When implementing features:

* Always prefer **clarity over completeness**
* Always think: “Would a DevOps engineer immediately understand this?”
* Avoid raw JSON unless explicitly toggled
* Use cards, spacing, and hierarchy
* Highlight impact:

  * size deltas
  * critical recommendations
  * regressions

DockSlim is a **product**, not a debugging tool.

---

## 8. Workflow Rules for AI Agents

1. Small, focused iterations (≈500–1000 LOC).
2. Do not refactor unrelated code.
3. Do not redesign UI unless requested.
4. Keep repository always runnable.
5. Always run:

   * `go test ./...`
6. Respect existing UX patterns.

---

## 9. Near-Term Roadmap (GUIDANCE)

Upcoming iterations should focus on:

1. Smart recommendations (heuristics-based)
2. Image comparison (diff between analyses)
3. Budgets / limits (fail on growth)
4. CI integrations (GitHub, GitLab)
5. Monetization:

   * advanced insights
   * history retention
   * exports

When in doubt, prioritize **features that increase perceived value**.
