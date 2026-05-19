-- set_updated_at is a shared trigger function that stamps updated_at to NOW()
-- on every UPDATE. Attach it to any table that has an updated_at column.
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;
