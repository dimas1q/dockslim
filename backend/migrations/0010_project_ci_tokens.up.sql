CREATE TABLE project_ci_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    token_hash TEXT NOT NULL,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ
);

CREATE INDEX idx_project_ci_tokens_project ON project_ci_tokens (project_id);
CREATE UNIQUE INDEX idx_project_ci_tokens_active_name ON project_ci_tokens (project_id, name) WHERE revoked_at IS NULL;
