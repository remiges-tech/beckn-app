-- Transactional outbox pattern for reliable async Beckn callbacks.
-- All outbound on_* calls are written here within the same DB transaction as
-- the state change. A background worker polls for PENDING events and delivers them.
-- Failed events are retried with exponential backoff per delivery_policy.
-- Events that exhaust retries are moved to dead_letter_events.

CREATE TABLE outbox_events (
    id                  UUID            PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Beckn addressing for the callback
    action              beckn_action    NOT NULL,
    transaction_id      UUID,
    message_id          UUID            NOT NULL DEFAULT gen_random_uuid(),

    -- Target callback URL
    target_url          TEXT            NOT NULL,
    bap_id              TEXT,
    bap_uri             TEXT,

    -- Payload to deliver
    payload             JSONB           NOT NULL,

    -- Delivery state machine
    status              outbox_status   NOT NULL DEFAULT 'PENDING',
    attempt_count       INTEGER         NOT NULL DEFAULT 0,
    max_attempts        INTEGER         NOT NULL DEFAULT 5,

    -- Exponential backoff: next_attempt_at = last_attempted_at * backoff_multiplier ^ attempt_count
    next_attempt_at     TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    last_attempted_at   TIMESTAMPTZ,
    last_error          TEXT,

    -- Optional link to the originating contract
    contract_id         UUID            REFERENCES contracts (id),

    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE outbox_events IS
    'Transactional outbox for reliable Beckn callback delivery. '
    'Written in the same transaction as state changes; consumed by the async delivery worker.';

COMMENT ON COLUMN outbox_events.next_attempt_at IS
    'Earliest time the worker should next attempt delivery. '
    'Backoff is applied by the worker after each failure.';

-- ---------------------------------------------------------------------------
-- Dead letter events — exhausted retry budget
-- ---------------------------------------------------------------------------

CREATE TABLE dead_letter_events (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Original outbox event preserved for manual inspection and replay
    outbox_event_id     UUID        NOT NULL,
    action              beckn_action NOT NULL,
    transaction_id      UUID,
    message_id          UUID,
    target_url          TEXT        NOT NULL,
    bap_id              TEXT,
    payload             JSONB       NOT NULL,
    total_attempts      INTEGER     NOT NULL,
    last_error          TEXT,
    contract_id         UUID        REFERENCES contracts (id),

    -- Human operator fields
    reviewed            BOOLEAN     NOT NULL DEFAULT FALSE,
    reviewed_at         TIMESTAMPTZ,
    reviewed_by         TEXT,
    resolution_note     TEXT,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE dead_letter_events IS
    'Beckn callbacks that exhausted the retry budget. '
    'Requires manual review and replay by the operations team.';

---- create above / drop below ----

DROP TABLE IF EXISTS dead_letter_events;
DROP TABLE IF EXISTS outbox_events;
