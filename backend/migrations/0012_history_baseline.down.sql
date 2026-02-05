DROP TABLE IF EXISTS project_policies;

DROP INDEX IF EXISTS idx_image_analyses_project_image_ref_analyzed_at;
DROP INDEX IF EXISTS idx_image_analyses_project_analyzed_at;

ALTER TABLE image_analyses
    DROP COLUMN IF EXISTS git_ref,
    DROP COLUMN IF EXISTS commit_sha,
    DROP COLUMN IF EXISTS analyzed_at,
    DROP COLUMN IF EXISTS layer_count,
    DROP COLUMN IF EXISTS largest_layer_bytes;
