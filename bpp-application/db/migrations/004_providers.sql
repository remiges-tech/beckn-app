-- Provider and provider location tables.
-- A Provider is the entity that offers resources (products/services) in a Catalog.
-- One BPP may host multiple providers (marketplace model).
-- provider_locations stores each physical site; geo is stored as plain JSONB (GeoJSON).

CREATE TABLE providers (
    id                          TEXT        NOT NULL,
    bpp_id                      TEXT        NOT NULL REFERENCES bpp_identities (bpp_id),

    -- Descriptor fields (flattened for query performance; also stored as JSONB below)
    descriptor_name             TEXT        NOT NULL,
    descriptor_code             TEXT,
    descriptor_short_desc       TEXT,
    descriptor_long_desc        TEXT,
    descriptor_thumbnail_image  TEXT,
    descriptor_docs             JSONB       NOT NULL DEFAULT '[]',
    descriptor_media_files      JSONB       NOT NULL DEFAULT '[]',

    -- JSON-LD domain-specific extension bag (Attributes schema)
    provider_attributes         JSONB,
    provider_attributes_context TEXT,
    provider_attributes_type    TEXT,

    is_active                   BOOLEAN     NOT NULL DEFAULT TRUE,

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at                  TIMESTAMPTZ,

    PRIMARY KEY (id, bpp_id)
);

COMMENT ON TABLE providers IS
    'Beckn Provider entities. Each provider belongs to one BPP and may appear in multiple catalogs.';

COMMENT ON COLUMN providers.provider_attributes IS
    'JSON-LD attributes bag for domain-specific provider extensions. Must include @context and @type.';

-- ---------------------------------------------------------------------------
-- Provider locations
-- Represents the availableAt[] array from the Provider schema.
-- geo is stored as plain JSONB (GeoJSON RFC 7946). Spatial filtering is
-- handled at the application layer or via jsonb operators.
-- ---------------------------------------------------------------------------

CREATE TABLE provider_locations (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

    provider_id         TEXT        NOT NULL,
    bpp_id              TEXT        NOT NULL,

    -- GeoJSON geometry stored verbatim (RFC 7946 — type + coordinates)
    geo                 JSONB       NOT NULL,

    -- Flattened address fields (schema.org PostalAddress) for plain SQL lookups
    address_street      TEXT,
    address_locality    TEXT,
    address_region      TEXT,
    address_country     TEXT,
    address_postal_code TEXT,
    address_extended    TEXT,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    FOREIGN KEY (provider_id, bpp_id) REFERENCES providers (id, bpp_id) ON DELETE CASCADE
);

COMMENT ON TABLE provider_locations IS
    'Physical locations where a provider operates. Maps to Provider.availableAt[].';

COMMENT ON COLUMN provider_locations.geo IS
    'Full GeoJSON geometry object (RFC 7946). Type and coordinates are stored verbatim.';

---- create above / drop below ----

DROP TABLE IF EXISTS provider_locations;
DROP TABLE IF EXISTS providers;
