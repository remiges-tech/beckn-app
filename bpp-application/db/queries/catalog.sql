-- Catalog publishing queries.
-- All operations use upsert or delete+insert to support FULL update mode.

-- ============================================================
-- PROVIDERS
-- ============================================================

-- name: UpsertProvider :exec
INSERT INTO providers (
    id, bpp_id,
    descriptor_name, descriptor_code, descriptor_short_desc, descriptor_long_desc,
    descriptor_docs, descriptor_media_files,
    is_active
) VALUES (
    $1, $2,
    $3, $4, $5, $6,
    $7, $8,
    true
)
ON CONFLICT (id, bpp_id) DO UPDATE SET
    descriptor_name      = EXCLUDED.descriptor_name,
    descriptor_code      = EXCLUDED.descriptor_code,
    descriptor_short_desc = EXCLUDED.descriptor_short_desc,
    descriptor_long_desc = EXCLUDED.descriptor_long_desc,
    descriptor_docs      = EXCLUDED.descriptor_docs,
    descriptor_media_files = EXCLUDED.descriptor_media_files,
    updated_at           = NOW();

-- name: DeleteProviderLocations :exec
DELETE FROM provider_locations
WHERE provider_id = $1 AND bpp_id = $2;

-- name: InsertProviderLocation :exec
INSERT INTO provider_locations (
    provider_id, bpp_id, geo,
    address_street, address_locality, address_region,
    address_country, address_postal_code, address_extended
) VALUES (
    $1, $2, $3,
    $4, $5, $6,
    $7, $8, $9
);

-- ============================================================
-- CATALOGS
-- ============================================================

-- name: UpsertCatalog :exec
INSERT INTO catalogs (
    id, bpp_id, bpp_uri, provider_id,
    descriptor_name, descriptor_code, descriptor_short_desc, descriptor_long_desc,
    descriptor_docs, descriptor_media_files,
    is_active, catalog_type, update_mode,
    validity_start, validity_end,
    version, schema_context
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7, $8,
    $9, $10,
    true, $11, 'FULL',
    $12, $13,
    1, '[]'
)
ON CONFLICT (id, bpp_id) DO UPDATE SET
    bpp_uri              = EXCLUDED.bpp_uri,
    provider_id          = EXCLUDED.provider_id,
    descriptor_name      = EXCLUDED.descriptor_name,
    descriptor_code      = EXCLUDED.descriptor_code,
    descriptor_short_desc = EXCLUDED.descriptor_short_desc,
    descriptor_long_desc = EXCLUDED.descriptor_long_desc,
    descriptor_docs      = EXCLUDED.descriptor_docs,
    descriptor_media_files = EXCLUDED.descriptor_media_files,
    catalog_type         = EXCLUDED.catalog_type,
    validity_start       = EXCLUDED.validity_start,
    validity_end         = EXCLUDED.validity_end,
    version              = catalogs.version + 1,
    updated_at           = NOW();

-- ============================================================
-- RESOURCES
-- ============================================================

-- name: DeleteCatalogResources :exec
DELETE FROM resources
WHERE catalog_id = $1 AND bpp_id = $2;

-- name: InsertResource :exec
INSERT INTO resources (
    id, catalog_id, bpp_id,
    descriptor_name, descriptor_code, descriptor_short_desc, descriptor_long_desc,
    descriptor_docs, descriptor_media_files,
    resource_attributes, resource_attributes_context, resource_attributes_type
) VALUES (
    $1, $2, $3,
    $4, $5, $6, $7,
    $8, $9,
    $10, $11, $12
);

-- ============================================================
-- OFFERS
-- ============================================================

-- name: DeleteCatalogOfferConsiderations :exec
DELETE FROM offer_considerations
WHERE catalog_id = $1 AND bpp_id = $2;

-- name: DeleteCatalogOffers :exec
DELETE FROM offers
WHERE catalog_id = $1 AND bpp_id = $2;

-- name: InsertOffer :exec
INSERT INTO offers (
    id, catalog_id, bpp_id,
    provider_id,
    descriptor_name, descriptor_code, descriptor_short_desc,
    descriptor_docs, descriptor_media_files,
    resource_ids,
    offer_attributes, offer_attributes_context, offer_attributes_type,
    validity_start, validity_end
) VALUES (
    $1, $2, $3,
    $4,
    $5, $6, $7,
    '[]', '[]',
    $8,
    $9, $10, $11,
    $12, $13
);

-- name: InsertOfferConsideration :exec
INSERT INTO offer_considerations (
    offer_id, catalog_id, bpp_id,
    consideration_id,
    status_code, status_name,
    consideration_attributes,
    consideration_attributes_context,
    consideration_attributes_type
) VALUES (
    $1, $2, $3,
    $4,
    $5, $6,
    $7,
    $8, $9
);
