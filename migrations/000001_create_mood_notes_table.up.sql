-- migrations/000001_create_mood_notes_table.up.sql
CREATE TABLE IF NOT EXISTS mood_notes (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- Add updated_at
    title TEXT NOT NULL CHECK (title <> ''),
    content TEXT NOT NULL CHECK (content <> ''),
    version INTEGER NOT NULL DEFAULT 1
);

-- Add a trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_mood_notes_updated_at
BEFORE UPDATE ON mood_notes
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();