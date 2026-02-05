ALTER TABLE image_analyses
    ADD COLUMN git_ref TEXT,
    ADD COLUMN commit_sha TEXT,
    ADD COLUMN analyzed_at TIMESTAMPTZ,
    ADD COLUMN layer_count INT,
    ADD COLUMN largest_layer_bytes BIGINT;

ALTER TABLE image_analyses
    ADD CONSTRAINT image_analyses_layer_count_check CHECK (layer_count >= 0),
    ADD CONSTRAINT image_analyses_largest_layer_check CHECK (largest_layer_bytes >= 0);

CREATE INDEX idx_image_analyses_project_analyzed_at
    ON image_analyses(project_id, analyzed_at DESC)
    WHERE status = 'completed';

CREATE INDEX idx_image_analyses_project_image_ref_analyzed_at
    ON image_analyses(project_id, image, git_ref, analyzed_at DESC)
    WHERE status = 'completed';

UPDATE image_analyses
SET analyzed_at = COALESCE(finished_at, created_at)
WHERE status = 'completed' AND analyzed_at IS NULL;

UPDATE image_analyses
SET layer_count = COALESCE(jsonb_array_length(result_json->'layers'), 0),
    largest_layer_bytes = COALESCE(
        (
            SELECT MAX((layer->>'size_bytes')::bigint)
            FROM jsonb_array_elements(COALESCE(result_json->'layers', '[]'::jsonb)) AS layer
        ),
        0
    )
WHERE status = 'completed'
  AND result_json IS NOT NULL
  AND (layer_count IS NULL OR largest_layer_bytes IS NULL);

CREATE TABLE project_policies (
    project_id UUID PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
    baseline_mode TEXT NOT NULL DEFAULT 'main_latest',
    baseline_ref_branch TEXT NOT NULL DEFAULT 'main',
    baseline_analysis_id UUID REFERENCES image_analyses(id) ON DELETE SET NULL,
    warn_delta_bytes BIGINT NULL CHECK (warn_delta_bytes >= 0),
    fail_delta_bytes BIGINT NULL CHECK (fail_delta_bytes >= 0),
    hard_limit_bytes BIGINT NULL CHECK (hard_limit_bytes >= 0),
    warn_delta_layers INT NULL CHECK (warn_delta_layers >= 0),
    fail_delta_layers INT NULL CHECK (fail_delta_layers >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO project_policies (project_id)
SELECT id FROM projects
ON CONFLICT DO NOTHING;
