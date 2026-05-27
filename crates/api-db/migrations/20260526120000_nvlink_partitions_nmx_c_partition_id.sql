ALTER TABLE nvlink_partitions
    ADD COLUMN nmx_c_partition_id INTEGER;

ALTER TABLE nvlink_partitions
    DROP CONSTRAINT IF EXISTS nvlink_partitions_nmx_m_id_key;

ALTER TABLE nvlink_partitions
    ALTER COLUMN nmx_m_id DROP NOT NULL;

ALTER TABLE nvlink_partitions
    ADD CONSTRAINT nvlink_partitions_external_id_check
    CHECK (
        nmx_m_id IS NOT NULL
        OR nmx_c_partition_id IS NOT NULL
    );
