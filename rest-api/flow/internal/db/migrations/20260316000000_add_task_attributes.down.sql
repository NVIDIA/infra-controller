ALTER TABLE task DROP COLUMN IF EXISTS attributes;
ALTER TABLE task ADD COLUMN IF NOT EXISTS component_uuids jsonb;
