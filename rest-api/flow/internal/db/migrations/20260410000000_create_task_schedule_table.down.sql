DROP INDEX IF EXISTS idx_task_schedule_scope_rack;
DROP TABLE IF EXISTS task_schedule_scope;
DROP TRIGGER IF EXISTS task_schedule_set_updated_at ON task_schedule;
DROP INDEX IF EXISTS idx_task_schedule_next_run;
DROP TABLE IF EXISTS task_schedule;
