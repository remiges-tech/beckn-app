-- Beckn message log: immutable audit trail of all inbound and outbound messages.
-- Partitioned by month on created_at for operational manageability.
-- Partitions must be created as the system runs; a maintenance job creates them ahead of time.
-- request_payload and response_payload store full JSON for non-repudiation.

CREATE TABLE beckn_message_log (
    id                      UUID            NOT NULL DEFAULT gen_random_uuid(),

    -- Beckn context fields
    transaction_id          UUID,
    message_id              UUID            NOT NULL,
    action                  beckn_action    NOT NULL,
    direction               message_direction NOT NULL,

    -- Network participants
    bap_id                  TEXT,
    bap_uri                 TEXT,
    bpp_id                  TEXT,
    bpp_uri                 TEXT,
    network_id              TEXT,
    domain                  TEXT,

    -- Response outcome
    ack_status              ack_status,

    -- Full payloads (stored for audit/replay; gzip compression via TOAST)
    request_payload         JSONB,
    response_payload        JSONB,

    -- Error details (when ack_status = NACK)
    error_code              TEXT,
    error_message           TEXT,

    -- Performance tracking
    processing_duration_ms  INTEGER,

    -- HTTP Signature of the inbound request (for non-repudiation)
    request_signature       TEXT,

    created_at              TIMESTAMPTZ     NOT NULL DEFAULT NOW(),

    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

COMMENT ON TABLE beckn_message_log IS
    'Immutable, partitioned audit log of all Beckn protocol messages. '
    'Provides non-repudiation — never update or delete rows.';

COMMENT ON COLUMN beckn_message_log.request_payload IS
    'Full JSON request body. Enables replay and dispute resolution.';

COMMENT ON COLUMN beckn_message_log.request_signature IS
    'Raw Authorization header value for signature re-verification.';

-- Create default partition for the current month on migration
-- The application/maintenance job creates future partitions ahead of schedule.
DO $$
DECLARE
    partition_start DATE := date_trunc('month', NOW())::DATE;
    partition_end   DATE := (date_trunc('month', NOW()) + INTERVAL '1 month')::DATE;
    partition_name  TEXT := 'beckn_message_log_' || to_char(partition_start, 'YYYY_MM');
BEGIN
    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF beckn_message_log '
        'FOR VALUES FROM (%L) TO (%L)',
        partition_name, partition_start, partition_end
    );
END $$;

---- create above / drop below ----

DROP TABLE IF EXISTS beckn_message_log CASCADE;
