CREATE TABLE image_analyses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    registry_id UUID REFERENCES registries(id) ON DELETE SET NULL,
    image TEXT NOT NULL,
    tag TEXT NOT NULL,
    status TEXT NOT NULL,
    total_size_bytes BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_image_analyses_project_created_at ON image_analyses(project_id, created_at DESC);
