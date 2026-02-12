ALTER TABLE users
ADD COLUMN IF NOT EXISTS is_admin BOOLEAN NOT NULL DEFAULT FALSE;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public'
          AND table_name = 'admin_users'
    ) THEN
        UPDATE users
        SET is_admin = TRUE
        WHERE id IN (SELECT user_id FROM admin_users);
    END IF;
END;
$$;

DROP TABLE IF EXISTS admin_users;
