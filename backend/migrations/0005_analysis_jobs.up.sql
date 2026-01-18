CREATE TABLE analysis_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    analysis_id UUID NOT NULL REFERENCES image_analyses(id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    attempts INT NOT NULL DEFAULT 0,
    locked_by TEXT,
    locked_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_analysis_jobs_status_created_at ON analysis_jobs(status, created_at);
CREATE INDEX idx_analysis_jobs_locked_at ON analysis_jobs(locked_at);

ALTER TABLE image_analyses
    ADD COLUMN result_json JSONB,
    ADD COLUMN started_at TIMESTAMPTZ,
    ADD COLUMN finished_at TIMESTAMPTZ;
