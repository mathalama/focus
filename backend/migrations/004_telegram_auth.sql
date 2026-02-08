CREATE TABLE IF NOT EXISTS telegram_link_codes (
    code TEXT PRIMARY KEY CHECK (code ~ '^[A-Z0-9]{8}$'),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_telegram_link_codes_user_id ON telegram_link_codes(user_id);
CREATE INDEX IF NOT EXISTS idx_telegram_link_codes_expires_at ON telegram_link_codes(expires_at);

CREATE TABLE IF NOT EXISTS telegram_identities (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    telegram_user_id BIGINT NOT NULL UNIQUE,
    telegram_username TEXT,
    telegram_first_name TEXT NOT NULL DEFAULT '',
    telegram_last_name TEXT NOT NULL DEFAULT '',
    linked_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
