-- BPP identity and cryptographic key registry.
-- Split into two tables:
--   bpp_identities  — canonical BPP subscriber record (bpp_id is the PK)
--   bpp_signing_keys — key rotation history; multiple active keys are allowed
--                      for zero-downtime rotation (old key remains valid during transition)
--
-- The Signature keyId format is: {bppId}|{uniqueKeyId}|ed25519
-- Verification: look up bpp_signing_keys by (bpp_id, key_id) and verify with public_key.

CREATE TABLE bpp_identities (
    bpp_id                      TEXT        PRIMARY KEY,

    bpp_uri                     TEXT        NOT NULL,

    -- Networks this BPP is registered on (dedi.global namespaces)
    network_ids                 TEXT[]      NOT NULL DEFAULT '{}',

    -- Domains served (retail, mobility, healthcare, energy, etc.)
    domains                     TEXT[]      NOT NULL DEFAULT '{}',

    is_active                   BOOLEAN     NOT NULL DEFAULT TRUE,

    registered_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE bpp_identities IS
    'Canonical BPP subscriber identities. bpp_id is the unique subscriber ID as registered on dedi.global.';

-- ---------------------------------------------------------------------------
-- BPP signing key management (supports key rotation)
-- ---------------------------------------------------------------------------

CREATE TABLE bpp_signing_keys (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

    bpp_id                  TEXT        NOT NULL REFERENCES bpp_identities (bpp_id) ON DELETE CASCADE,

    -- Unique key identifier used in Signature: {bppId}|{key_id}|ed25519
    key_id                  TEXT        NOT NULL,

    -- Ed25519 public key (base64-encoded)
    signing_public_key       TEXT        NOT NULL,

    -- Optional X25519 encryption public key
    encryption_public_key    TEXT,

    is_active               BOOLEAN     NOT NULL DEFAULT TRUE,
    valid_from              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    valid_until             TIMESTAMPTZ,

    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT bpp_signing_keys_bpp_key_unique UNIQUE (bpp_id, key_id)
);

COMMENT ON TABLE bpp_signing_keys IS
    'Ed25519 signing key registry for each BPP. Supports zero-downtime key rotation. '
    'Multiple keys may be active simultaneously during transition windows.';

COMMENT ON COLUMN bpp_signing_keys.key_id IS
    'The uniqueKeyId segment in the Signature keyId: {bppId}|{uniqueKeyId}|ed25519';

---- create above / drop below ----

DROP TABLE IF EXISTS bpp_signing_keys;
DROP TABLE IF EXISTS bpp_identities;
