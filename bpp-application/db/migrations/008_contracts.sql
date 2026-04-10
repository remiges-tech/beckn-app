-- Contracts: the canonical transaction object in Beckn v2.0.
-- Replaces the v1 Order construct. Covers retail, mobility, hiring, healthcare,
-- energy, carbon trading, logistics, financial services, and more.
-- One contract spans the full lifecycle: DRAFT → ACTIVE → COMPLETE | CANCELLED.

CREATE TABLE contracts (
    id                          UUID        PRIMARY KEY,

    -- Beckn network addressing (from Context)
    bpp_id                      TEXT        NOT NULL,
    bap_id                      TEXT        NOT NULL,
    bap_uri                     TEXT        NOT NULL,

    -- transaction_id persists across all API calls for a single user session
    transaction_id              UUID        NOT NULL,

    -- Network and domain context
    network_id                  TEXT,
    domain                      TEXT,

    -- Contract lifecycle
    status                      contract_status NOT NULL DEFAULT 'DRAFT',

    -- Descriptor (human/agent-readable description of the contract type)
    descriptor_name             TEXT,
    descriptor_code             TEXT,
    descriptor_short_desc       TEXT,
    descriptor_long_desc        TEXT,
    descriptor_docs             JSONB       NOT NULL DEFAULT '[]',
    descriptor_media_files      JSONB       NOT NULL DEFAULT '[]',

    -- JSON-LD domain-specific extension bag (contractAttributes)
    contract_attributes         JSONB,
    contract_attributes_context TEXT,
    contract_attributes_type    TEXT,

    -- Soft delete — cancelled/completed contracts are archived, not removed
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at                  TIMESTAMPTZ,

    CONSTRAINT contracts_bpp_id_fk FOREIGN KEY (bpp_id)
        REFERENCES bpp_identities (bpp_id)
);

COMMENT ON TABLE contracts IS
    'Beckn Contract — the central transaction object binding commitments, consideration, '
    'performance, settlements, and participants across the full protocol lifecycle.';

COMMENT ON COLUMN contracts.transaction_id IS
    'Persists across discover→confirm for a single user session (from Context.transactionId).';

COMMENT ON COLUMN contracts.contract_attributes IS
    'JSON-LD extension attributes for domain-specific contract data. '
    'Must carry @context and @type when populated.';

---- create above / drop below ----

DROP TABLE IF EXISTS contracts;
