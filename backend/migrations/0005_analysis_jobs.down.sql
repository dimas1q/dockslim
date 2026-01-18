ALTER TABLE image_analyses
    DROP COLUMN IF EXISTS result_json,
    DROP COLUMN IF EXISTS started_at,
    DROP COLUMN IF EXISTS finished_at;

DROP TABLE IF EXISTS analysis_jobs;
