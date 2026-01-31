CREATE TABLE project_budgets (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    image TEXT NULL,
    warn_delta_bytes BIGINT NULL CHECK (warn_delta_bytes >= 0),
    fail_delta_bytes BIGINT NULL CHECK (fail_delta_bytes >= 0),
    hard_limit_bytes BIGINT NULL CHECK (hard_limit_bytes >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (image IS NULL OR length(trim(image)) > 0)
);

ALTER TABLE project_budgets
ADD CONSTRAINT project_budgets_project_image_unique UNIQUE (project_id, image);

CREATE UNIQUE INDEX project_budgets_default_unique ON project_budgets(project_id) WHERE image IS NULL;
