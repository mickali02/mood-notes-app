-- migrations/000001_create_mood_notes_table.down.sql
DROP TRIGGER IF EXISTS update_mood_notes_updated_at ON mood_notes;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS mood_notes;