-- Transactions table: one row per Beckn transactionId.
-- Tracks the overall lifecycle of a buyer session.

CREATE TABLE transactions (
    transaction_id  UUID            NOT NULL PRIMARY KEY,
    bap_id          TEXT            NOT NULL REFERENCES bap_identities(bap_id),
    network_id      TEXT,
    bpp_id          TEXT,
    bpp_uri         TEXT,
    status          transaction_status  NOT NULL DEFAULT 'SELECT_SENT',
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE transactions IS
    'One row per Beckn transactionId initiated by this BAP. '
    'Status advances as select → init → confirm actions progress.';

---- create above / drop below ----

DROP TABLE IF EXISTS transactions;
