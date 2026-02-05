DROP INDEX IF EXISTS idx_image_analyses_project_analyzed_at;
DROP INDEX IF EXISTS idx_image_analyses_project_image_ref_analyzed_at;

CREATE INDEX idx_image_analyses_project_analyzed_at
    ON image_analyses(project_id, analyzed_at DESC);

CREATE INDEX idx_image_analyses_project_image_ref_analyzed_at
    ON image_analyses(project_id, image, git_ref, analyzed_at DESC);

CREATE INDEX idx_image_analyses_project_status_analyzed_at
    ON image_analyses(project_id, status, analyzed_at DESC);
