-- Domain enums for the BAP application (mirrors bpp enums).

-- All Beckn protocol actions (context.action)
CREATE TYPE beckn_action AS ENUM (
    'discover',
    'on_discover',
    'select',
    'on_select',
    'init',
    'on_init',
    'confirm',
    'on_confirm',
    'status',
    'on_status',
    'track',
    'on_track',
    'update',
    'on_update',
    'cancel',
    'on_cancel',
    'rate',
    'on_rate',
    'support',
    'on_support'
);

-- Synchronous response acknowledgement
CREATE TYPE ack_status AS ENUM (
    'ACK',
    'NACK'
);

-- Message flow direction (BAP perspective)
CREATE TYPE message_direction AS ENUM (
    'INBOUND',   -- BAP received a callback from BPP (on_select, on_init …)
    'OUTBOUND'   -- BAP sent a request to BPP (select, init …)
);

-- Beckn transaction lifecycle (from BAP perspective)
CREATE TYPE transaction_status AS ENUM (
    'SELECT_SENT',
    'QUOTE_RECEIVED',    -- on_select received
    'INIT_SENT',
    'INIT_RECEIVED',     -- on_init received
    'CONFIRM_SENT',
    'CONFIRMED',         -- on_confirm received (ACTIVE)
    'CANCELLED',
    'COMPLETE'
);

---- create above / drop below ----

DROP TYPE IF EXISTS transaction_status;
DROP TYPE IF EXISTS message_direction;
DROP TYPE IF EXISTS ack_status;
DROP TYPE IF EXISTS beckn_action;
