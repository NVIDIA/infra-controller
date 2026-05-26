-- Add started_at to record when a task actually begins execution.
ALTER TABLE task ADD COLUMN IF NOT EXISTS started_at TIMESTAMPTZ;
