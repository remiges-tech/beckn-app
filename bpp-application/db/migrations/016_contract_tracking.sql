-- Contract tracking sessions.
-- Each row is a live tracking handle returned via /on_track.
-- url is the off-network tracking endpoint (HTTP URL or WebSocket).
-- tracking_attributes carries domain-specific context (route, ETA, vehicle, etc.).
-- Sessions expire via expires_at and are invalidated by setting status = INACTIVE.

CREATE TABLE contract_tracking (
    id                          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

    contract_id                 UUID        NOT NULL REFERENCES contracts (id) ON DELETE CASCADE,

    -- Performance unit being tracked (optional link)
    performance_id              TEXT,

    status                      tracking_status NOT NULL DEFAULT 'ACTIVE',

    -- Off-network tracking URL / WebSocket endpoint
    url                         TEXT        NOT NULL,

    -- Domain-specific tracking context (JSON-LD Attributes)
    tracking_attributes         JSONB,
    tracking_attributes_context TEXT,
    tracking_attributes_type    TEXT,

    expires_at                  TIMESTAMPTZ,

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE contract_tracking IS
    'Real-time tracking handles for active contract fulfillment. '
    'Not for contract state changes (use on_status for that).';

---- create above / drop below ----

DROP TABLE IF EXISTS contract_tracking;
