-- Contract snapshots: stores the full contract JSON received from the BPP
-- for each on_* callback. This gives the BAP a complete audit trail.
-- One row per (transaction_id, action). Upserting on (transaction_id, action)
-- means a retried on_init simply overwrites the previous snapshot.

CREATE TABLE contract_snapshots (
    id              UUID            NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    transaction_id  UUID            NOT NULL REFERENCES transactions(transaction_id),
    action          beckn_action    NOT NULL,  -- 'on_select' | 'on_init' | 'on_confirm' etc.
    contract_id     TEXT,                       -- The id the BPP assigned to the contract
    contract        JSONB           NOT NULL,   -- Full contract object as received
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_snapshot_txn_action UNIQUE (transaction_id, action)
);

COMMENT ON TABLE contract_snapshots IS
    'Stores the full contract JSON received from the BPP at each on_* callback. '
    'The BAP uses this to track the evolving contract state through the Beckn lifecycle.';

COMMENT ON COLUMN contract_snapshots.contract_id IS
    'The UUID the BPP assigned to the contract (populated from on_init onwards).';

---- create above / drop below ----

DROP TABLE IF EXISTS contract_snapshots;
