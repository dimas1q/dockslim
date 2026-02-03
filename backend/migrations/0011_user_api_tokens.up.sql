CREATE TABLE user_api_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    token_hash TEXT NOT NULL,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ
);

CREATE INDEX idx_user_api_tokens_user_id ON user_api_tokens(user_id);
CREATE UNIQUE INDEX idx_user_api_tokens_user_name_active ON user_api_tokens(user_id, name) WHERE revoked_at IS NULL;
