-- Catalogs published by the BPP to Catalog Discovery Services (CDS).
-- Versioning is achieved via version INTEGER — each publish increments it.
-- validity_* bounds control when a catalog is live.

CREATE TABLE catalogs (
    id                          TEXT        NOT NULL,
    bpp_id                      TEXT        NOT NULL REFERENCES bpp_identities (bpp_id),
    bpp_uri                     TEXT        NOT NULL,

    provider_id                 TEXT        NOT NULL,

    -- Descriptor (human/agent-readable catalog description)
    descriptor_name             TEXT        NOT NULL,
    descriptor_code             TEXT,
    descriptor_short_desc       TEXT,
    descriptor_long_desc        TEXT,
    descriptor_thumbnail_image  TEXT,
    descriptor_docs             JSONB       NOT NULL DEFAULT '[]',
    descriptor_media_files      JSONB       NOT NULL DEFAULT '[]',

    is_active                   BOOLEAN     NOT NULL DEFAULT TRUE,

    -- publishDirectives
    catalog_type                catalog_type,
    update_mode                 catalog_update_mode NOT NULL DEFAULT 'FULL',

    -- TimePeriod validity window
    validity_start              TIMESTAMPTZ,
    validity_end                TIMESTAMPTZ,

    -- Monotonically increasing version per catalog ID
    version                     INTEGER     NOT NULL DEFAULT 1,

    -- JSON-LD schema context array from Context.schemaContext
    schema_context              JSONB       NOT NULL DEFAULT '[]',

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at                  TIMESTAMPTZ,

    PRIMARY KEY (id, bpp_id),

    FOREIGN KEY (provider_id, bpp_id) REFERENCES providers (id, bpp_id)
);

COMMENT ON TABLE catalogs IS
    'Beckn Catalog objects published by the BPP. Versioned — each publish call increments version.';

COMMENT ON COLUMN catalogs.schema_context IS
    'Array of JSON-LD context URIs for the schemas used by resources in this catalog.';

-- ---------------------------------------------------------------------------
-- Catalog publish results (CDS processing responses)
-- Maps to CatalogProcessingResult from CatalogOnPublishAction
-- ---------------------------------------------------------------------------

CREATE TABLE catalog_publish_results (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    catalog_id      TEXT        NOT NULL,
    bpp_id          TEXT        NOT NULL,
    catalog_version INTEGER     NOT NULL,

    status          TEXT        NOT NULL CHECK (status IN ('ACCEPTED', 'REJECTED', 'PARTIAL')),
    errors          JSONB       NOT NULL DEFAULT '[]',
    stats           JSONB,

    published_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    FOREIGN KEY (catalog_id, bpp_id) REFERENCES catalogs (id, bpp_id)
);

COMMENT ON TABLE catalog_publish_results IS
    'Processing results returned by the CDS for each catalog publish attempt.';

---- create above / drop below ----

DROP TABLE IF EXISTS catalog_publish_results;
DROP TABLE IF EXISTS catalogs;
