-- Add Strict Mode to Focus Sessions
ALTER TABLE focus_sessions ADD COLUMN IF NOT EXISTS is_strict BOOLEAN NOT NULL DEFAULT FALSE;

-- Add Tags to Goals
ALTER TABLE goals ADD COLUMN IF NOT EXISTS tags TEXT[] DEFAULT '{}';

-- Gamification: Shop Items
CREATE TABLE IF NOT EXISTS items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL,
    cost INT NOT NULL CHECK (cost >= 0),
    type TEXT NOT NULL, -- 'bee', 'decoration', 'sound', 'theme'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Gamification: User Inventory
CREATE TABLE IF NOT EXISTS user_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    purchased_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, item_id)
);

-- Seed some initial items
INSERT INTO items (name, description, cost, type) VALUES
('Worker Bee', 'A diligent worker for your hive.', 50, 'bee'),
('Queen Bee', 'The heart of the colony.', 500, 'bee'),
('Golden Honeycomb', 'A shiny decoration for your hive.', 200, 'decoration'),
('Rainforest Sound', 'Soothing sounds of rain.', 100, 'sound')
ON CONFLICT (name) DO NOTHING;
