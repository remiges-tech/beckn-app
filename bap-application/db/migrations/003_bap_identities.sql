-- BAP subscriber identity registry.
-- Seeded once on startup from environment variables.

CREATE TABLE bap_identities (
    bap_id      TEXT        NOT NULL PRIMARY KEY,
    bap_uri     TEXT        NOT NULL,
    network_ids TEXT[]      NOT NULL DEFAULT '{}',
    is_active   BOOLEAN     NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE bap_identities IS
    'Registered BAP subscribers. Seeded at startup from BAP_ID / BAP_URI env vars.';

---- create above / drop below ----

DROP TABLE IF EXISTS bap_identities;
