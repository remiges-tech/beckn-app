-- Settlements: records representing the discharge of agreed consideration.
-- Each settlement references a consideration_id from the same contract.
-- Tracks progression from DRAFT → COMMITTED → COMPLETE.

CREATE TABLE settlements (
    id                          TEXT        NOT NULL,
    contract_id                 UUID        NOT NULL REFERENCES contracts (id) ON DELETE CASCADE,

    -- Which consideration this settlement discharges
    consideration_id            TEXT        NOT NULL,

    status                      settlement_status NOT NULL DEFAULT 'DRAFT',

    -- JSON-LD attributes bag (payment gateway refs, UPI IDs, bank details, etc.)
    settlement_attributes       JSONB,
    settlement_attributes_context TEXT,
    settlement_attributes_type  TEXT,

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (id, contract_id)
);

COMMENT ON TABLE settlements IS
    'Discharge records for consideration agreed in a Contract. '
    'Carries payment gateway references or other proof of value transfer.';

---- create above / drop below ----

DROP TABLE IF EXISTS settlements;
