CREATE TABLE IF NOT EXISTS product_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    event_name TEXT NOT NULL,
    source TEXT NOT NULL DEFAULT 'api',
    properties JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_product_events_user_id
    ON product_events(user_id);

CREATE INDEX IF NOT EXISTS idx_product_events_event_name
    ON product_events(event_name);

CREATE INDEX IF NOT EXISTS idx_product_events_created_at
    ON product_events(created_at DESC);
