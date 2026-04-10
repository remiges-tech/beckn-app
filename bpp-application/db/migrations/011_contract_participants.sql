-- Contract participants: all parties involved in a contract.
-- May include the consumer, provider, logistics agent, third-party guarantor, etc.
-- Each participant carries a role expressed via participantAttributes @type.

CREATE TABLE contract_participants (
    id                              TEXT        NOT NULL,
    contract_id                     UUID        NOT NULL REFERENCES contracts (id) ON DELETE CASCADE,

    -- Descriptor (name/code of the participant)
    descriptor_name                 TEXT,
    descriptor_code                 TEXT,
    descriptor_short_desc           TEXT,
    descriptor_long_desc            TEXT,

    -- JSON-LD extension bag (carries role, contact info, credentials, etc.)
    participant_attributes          JSONB,
    participant_attributes_context  TEXT,
    participant_attributes_type     TEXT,

    created_at                      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (id, contract_id)
);

COMMENT ON TABLE contract_participants IS
    'Parties involved in a Contract. Role is expressed via participant_attributes @type.';

---- create above / drop below ----

DROP TABLE IF EXISTS contract_participants;
