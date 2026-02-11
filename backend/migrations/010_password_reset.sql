-- Add password_reset_tokens table for password reset functionality
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_id
    ON password_reset_tokens(user_id);

CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_expires_at
    ON password_reset_tokens(expires_at);

-- Add email type to outbox to support different email types
ALTER TABLE email_outbox ADD COLUMN IF NOT EXISTS email_type TEXT DEFAULT 'verification';
ALTER TABLE email_outbox ADD COLUMN IF NOT EXISTS reset_link TEXT;
