DROP INDEX IF EXISTS public.firmware_update_state_created_idx;
DROP INDEX IF EXISTS public.firmware_update_created_at_idx;
DROP INDEX IF EXISTS public.firmware_update_state_idx;
DROP TABLE IF EXISTS public.firmware_update;

DROP INDEX IF EXISTS public.pmc_vendor_idx;
DROP TABLE IF EXISTS public.pmc;
