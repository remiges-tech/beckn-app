-- Beckn message log: immutable audit trail of all inbound on_* callbacks
-- and any outbound Beckn requests. Mirrors bpp beckn_message_log schema.
-- Partitioned by month for manageability.

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

    -- URL that was called (INBOUND = our endpoint; OUTBOUND = remote endpoint)
    url                     TEXT,

    -- Response outcome
    ack_status              ack_status,

    -- Full payloads (stored for audit/replay)
    request_payload         JSONB,
    response_payload        JSONB,

    -- Error details (when ack_status = NACK)
    error_message           TEXT,

    -- Performance tracking
    processing_duration_ms  INTEGER,

    created_at              TIMESTAMPTZ     NOT NULL DEFAULT NOW(),

    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

COMMENT ON TABLE beckn_message_log IS
    'Immutable audit log of all Beckn protocol messages received or sent by this BAP.';

-- Create default partition for current month
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
