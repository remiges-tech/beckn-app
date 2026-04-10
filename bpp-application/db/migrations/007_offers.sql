-- Offers: pricing and terms under which resources may be committed.
-- An offer references one or more resources by ID and carries considerations
-- (price/value) and optional visibility constraints (availableTo).
-- offer_considerations and offer_available_to are child tables to allow clean querying.

CREATE TABLE offers (
    id                          TEXT        NOT NULL,
    catalog_id                  TEXT        NOT NULL,
    bpp_id                      TEXT        NOT NULL,

    -- Provider is optional at the offer level (may differ from catalog provider)
    provider_id                 TEXT,

    -- Descriptor
    descriptor_name             TEXT,
    descriptor_code             TEXT,
    descriptor_short_desc       TEXT,
    descriptor_long_desc        TEXT,
    descriptor_thumbnail_image  TEXT,
    descriptor_docs             JSONB       NOT NULL DEFAULT '[]',
    descriptor_media_files      JSONB       NOT NULL DEFAULT '[]',

    -- Resource IDs covered by this offer (denormalized for fast lookup)
    resource_ids                TEXT[]      NOT NULL DEFAULT '{}',

    -- Add-ons: array of embedded Resource or Offer objects
    add_ons                     JSONB       NOT NULL DEFAULT '[]',

    -- JSON-LD domain-specific extension bag
    offer_attributes            JSONB,
    offer_attributes_context    TEXT,
    offer_attributes_type       TEXT,

    -- Validity window
    validity_start              TIMESTAMPTZ,
    validity_end                TIMESTAMPTZ,

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at                  TIMESTAMPTZ,

    PRIMARY KEY (id, catalog_id, bpp_id),

    FOREIGN KEY (catalog_id, bpp_id) REFERENCES catalogs (id, bpp_id) ON DELETE CASCADE
);

COMMENT ON TABLE offers IS
    'Beckn Offer entities. Defines the terms and value proposition for committing one or more resources.';

-- ---------------------------------------------------------------------------
-- Offer considerations (proposed value to be exchanged)
-- Maps to Offer.considerations[] → Consideration
-- ---------------------------------------------------------------------------

CREATE TABLE offer_considerations (
    id                                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

    offer_id                            TEXT        NOT NULL,
    catalog_id                          TEXT        NOT NULL,
    bpp_id                              TEXT        NOT NULL,

    consideration_id                    TEXT        NOT NULL,

    -- Status is a Descriptor in the spec; store code + name
    status_code                         TEXT        NOT NULL,
    status_name                         TEXT,

    -- JSON-LD attributes (PriceSpecification for monetary, or generic Attributes)
    consideration_attributes            JSONB,
    consideration_attributes_context    TEXT,
    consideration_attributes_type       TEXT,

    FOREIGN KEY (offer_id, catalog_id, bpp_id) REFERENCES offers (id, catalog_id, bpp_id) ON DELETE CASCADE
);

COMMENT ON TABLE offer_considerations IS
    'Proposed value exchange terms attached to an Offer (price, credits, service exchange, etc.).';

-- ---------------------------------------------------------------------------
-- Offer visibility constraints (availableTo)
-- Controls which network participants can discover or transact on an offer.
-- ---------------------------------------------------------------------------

CREATE TABLE offer_available_to (
    id          UUID    PRIMARY KEY DEFAULT gen_random_uuid(),

    offer_id    TEXT    NOT NULL,
    catalog_id  TEXT    NOT NULL,
    bpp_id      TEXT    NOT NULL,

    -- Constraint type: NETWORK | PARTICIPANT | ROLE
    scope_type  TEXT    NOT NULL CHECK (scope_type IN ('NETWORK', 'PARTICIPANT', 'ROLE')),
    scope_id    TEXT    NOT NULL,

    FOREIGN KEY (offer_id, catalog_id, bpp_id) REFERENCES offers (id, catalog_id, bpp_id) ON DELETE CASCADE
);

COMMENT ON TABLE offer_available_to IS
    'Visibility constraints on an Offer. When absent the offer is visible to all network participants.';

---- create above / drop below ----

DROP TABLE IF EXISTS offer_available_to;
DROP TABLE IF EXISTS offer_considerations;
DROP TABLE IF EXISTS offers;
