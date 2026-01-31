ALTER TABLE users ADD COLUMN login TEXT;

UPDATE users
SET login = email
WHERE login IS NULL;

ALTER TABLE users ALTER COLUMN login SET NOT NULL;

CREATE UNIQUE INDEX idx_users_login_unique ON users (login);
