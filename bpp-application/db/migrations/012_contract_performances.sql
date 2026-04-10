-- Contract performances (fulfillment execution units).
-- Performance is the generalized fulfillment abstraction in Beckn v2.0.
-- Each performance unit represents one execution plan:
--   physical delivery, service provisioning, API access, subscription activation,
--   carbon credit transfer, capacity allocation, workforce onboarding, etc.
-- Domain-specific delivery details live in performance_attributes (JSON-LD).

CREATE TABLE contract_performances (
    id                                  TEXT        NOT NULL,
    contract_id                         UUID        NOT NULL REFERENCES contracts (id) ON DELETE CASCADE,

    -- Status expressed as a Descriptor (code + name)
    status_code                         TEXT,
    status_name                         TEXT,
    status_short_desc                   TEXT,

    -- References to commitments being fulfilled by this performance unit
    commitment_ids                      TEXT[]      NOT NULL DEFAULT '{}',

    -- JSON-LD domain-specific details
    -- For hyperlocal delivery: use beckn:HyperlocalDelivery (aligned with schema:ParcelDelivery)
    -- For other types: use generic Attributes
    performance_attributes              JSONB,
    performance_attributes_context      TEXT,
    performance_attributes_type         TEXT,

    created_at                          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (id, contract_id)
);

COMMENT ON TABLE contract_performances IS
    'Generalized fulfillment execution units for a Contract. '
    'Each row represents one delivery/service/activation plan. '
    'Tracking and support interactions link here.';

COMMENT ON COLUMN contract_performances.commitment_ids IS
    'IDs of commitments being executed by this performance unit.';

---- create above / drop below ----

DROP TABLE IF EXISTS contract_performances;
