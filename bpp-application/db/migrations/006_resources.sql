-- Resources: domain-neutral units of value within a catalog.
-- Covers retail SKUs, mobility rides, job roles, carbon credits, dataset entitlements, etc.
-- search_vector is maintained by a trigger (migration 021) for full-text discovery.

CREATE TABLE resources (
    id                          TEXT        NOT NULL,
    catalog_id                  TEXT        NOT NULL,
    bpp_id                      TEXT        NOT NULL,

    -- Descriptor (human/agent-readable item description)
    descriptor_name             TEXT,
    descriptor_code             TEXT,
    descriptor_short_desc       TEXT,
    descriptor_long_desc        TEXT,
    descriptor_thumbnail_image  TEXT,
    descriptor_docs             JSONB       NOT NULL DEFAULT '[]',
    descriptor_media_files      JSONB       NOT NULL DEFAULT '[]',

    -- JSON-LD domain-specific attributes bag (Attributes schema)
    resource_attributes         JSONB,
    resource_attributes_context TEXT,
    resource_attributes_type    TEXT,

    -- For regular catalogs: pointer to the master resource being extended
    master_resource_id          TEXT,

    -- Weighted full-text search vector (maintained by trigger)
    search_vector               TSVECTOR,

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at                  TIMESTAMPTZ,

    PRIMARY KEY (id, catalog_id, bpp_id),

    FOREIGN KEY (catalog_id, bpp_id) REFERENCES catalogs (id, bpp_id) ON DELETE CASCADE
);

COMMENT ON TABLE resources IS
    'Beckn Resource entities. Domain-neutral unit of value discoverable and committable on the network.';

COMMENT ON COLUMN resources.resource_attributes IS
    'JSON-LD attributes bag. MUST contain @context and @type per the Beckn Attributes schema.';

COMMENT ON COLUMN resources.search_vector IS
    'PostgreSQL tsvector for full-text search across name, shortDesc, longDesc, and attributes. '
    'Updated automatically by trigger on insert/update.';

---- create above / drop below ----

DROP TABLE IF EXISTS resources;
