-- Considerations: value agreed to be exchanged under a Contract.
-- Domain-neutral — covers monetary price, tokens/credits, asset transfers,
-- service exchanges, and compliance artifacts.

CREATE TABLE considerations (
    id                                  TEXT        NOT NULL,
    contract_id                         UUID        NOT NULL REFERENCES contracts (id) ON DELETE CASCADE,

    -- Status expressed as a Descriptor (code + name)
    status_code                         TEXT        NOT NULL,
    status_name                         TEXT,

    -- JSON-LD attributes bag
    -- For monetary value: use PriceSpecification schema (@type: PriceSpecification)
    -- For other types: use generic Attributes
    consideration_attributes            JSONB,
    consideration_attributes_context    TEXT,
    consideration_attributes_type       TEXT,

    created_at                          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (id, contract_id)
);

COMMENT ON TABLE considerations IS
    'Value promised to be exchanged under a Contract. '
    'Use PriceSpecification attributes for monetary; generic Attributes for other types.';

---- create above / drop below ----

DROP TABLE IF EXISTS considerations;
