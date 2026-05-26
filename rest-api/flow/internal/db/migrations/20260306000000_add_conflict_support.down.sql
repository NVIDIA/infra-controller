DROP INDEX IF EXISTS idx_task_rack_status;
ALTER TABLE task DROP COLUMN IF EXISTS queue_expires_at;
