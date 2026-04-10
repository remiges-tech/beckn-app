-- Commitments, their resources, and their offers.
-- A Commitment is a structured sub-agreement within a Contract.
-- Each commitment references exactly one Offer and one or more Resources with quantities.

CREATE TABLE commitments (
    id                          TEXT        NOT NULL,
    contract_id                 UUID        NOT NULL REFERENCES contracts (id) ON DELETE CASCADE,

    status                      commitment_status NOT NULL DEFAULT 'DRAFT',

    -- JSON-LD domain-specific extension bag
    commitment_attributes       JSONB,
    commitment_attributes_context TEXT,
    commitment_attributes_type  TEXT,

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (id, contract_id)
);

COMMENT ON TABLE commitments IS
    'Structured sub-agreements within a Contract. Each commitment binds resources to an offer.';

-- ---------------------------------------------------------------------------
-- Resources committed within a commitment
-- Maps to Commitment.resources[] where each item requires id + quantity
-- ---------------------------------------------------------------------------

CREATE TABLE commitment_resources (
    id                          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

    commitment_id               TEXT        NOT NULL,
    contract_id                 UUID        NOT NULL,

    -- Resource identity (denormalized — resource may have changed since commit)
    resource_id                 TEXT        NOT NULL,
    resource_descriptor         JSONB,
    resource_attributes         JSONB,

    -- Quantity: JSONB to accommodate domain-specific measure schemas
    -- (e.g. {value: 2, unit: "pieces"} or {measure: "kWh", value: 50})
    quantity                    JSONB       NOT NULL,

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    FOREIGN KEY (commitment_id, contract_id) REFERENCES commitments (id, contract_id) ON DELETE CASCADE
);

COMMENT ON TABLE commitment_resources IS
    'Resources selected within a commitment. Quantity is domain-agnostic JSONB.';

-- ---------------------------------------------------------------------------
-- Offer selected within a commitment
-- A commitment has exactly one Offer (1:1 per spec)
-- ---------------------------------------------------------------------------

CREATE TABLE commitment_offers (
    id                          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

    commitment_id               TEXT        NOT NULL,
    contract_id                 UUID        NOT NULL,

    -- Offer identity (denormalized snapshot at selection time)
    offer_id                    TEXT        NOT NULL,
    offer_descriptor            JSONB,
    offer_attributes            JSONB,

    -- Resource IDs covered by the offer (snapshot)
    resource_ids                TEXT[]      NOT NULL DEFAULT '{}',

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (commitment_id, contract_id),

    FOREIGN KEY (commitment_id, contract_id) REFERENCES commitments (id, contract_id) ON DELETE CASCADE
);

COMMENT ON TABLE commitment_offers IS
    'The single Offer selected for a commitment. Stored as a snapshot to preserve agreed terms.';

---- create above / drop below ----

DROP TABLE IF EXISTS commitment_offers;
DROP TABLE IF EXISTS commitment_resources;
DROP TABLE IF EXISTS commitments;
