-- Production indexes.
-- Organized by table. Index naming convention: idx_{table}_{columns}.
-- GIN indexes cover JSONB and array columns (uses core PostgreSQL GIN, no extensions needed).
-- BRIN indexes cover append-only timestamp columns on large tables.

-- ---------------------------------------------------------------------------
-- bpp_identities + bpp_signing_keys
-- ---------------------------------------------------------------------------

CREATE INDEX idx_bpp_identities_active
    ON bpp_identities (bpp_id, is_active)
    WHERE is_active = TRUE;

CREATE INDEX idx_bpp_identities_network_ids
    ON bpp_identities USING GIN (network_ids);

-- Fast key lookup during Signature verification: O(log n) by (bpp_id, key_id)
CREATE INDEX idx_bpp_signing_keys_lookup
    ON bpp_signing_keys (bpp_id, key_id, is_active)
    WHERE is_active = TRUE;

-- ---------------------------------------------------------------------------
-- providers
-- ---------------------------------------------------------------------------

CREATE INDEX idx_providers_bpp_id
    ON providers (bpp_id);

CREATE INDEX idx_providers_active
    ON providers (bpp_id, is_active)
    WHERE is_active = TRUE AND deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- provider_locations
-- ---------------------------------------------------------------------------

CREATE INDEX idx_provider_locations_provider
    ON provider_locations (provider_id, bpp_id);

-- GIN index for jsonb containment queries on the geo field (e.g. filter by type)
CREATE INDEX idx_provider_locations_geo
    ON provider_locations USING GIN (geo jsonb_path_ops);

-- B-tree indexes for address-based filtering
CREATE INDEX idx_provider_locations_locality
    ON provider_locations (address_locality)
    WHERE address_locality IS NOT NULL;

CREATE INDEX idx_provider_locations_country
    ON provider_locations (address_country)
    WHERE address_country IS NOT NULL;

-- ---------------------------------------------------------------------------
-- catalogs
-- ---------------------------------------------------------------------------

CREATE INDEX idx_catalogs_bpp_active
    ON catalogs (bpp_id, is_active)
    WHERE is_active = TRUE AND deleted_at IS NULL;

CREATE INDEX idx_catalogs_provider
    ON catalogs (provider_id, bpp_id);

CREATE INDEX idx_catalogs_validity
    ON catalogs (validity_start, validity_end)
    WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- resources
-- ---------------------------------------------------------------------------

-- Full-text search (GIN on tsvector maintained by trigger)
CREATE INDEX idx_resources_search_vector
    ON resources USING GIN (search_vector);

-- JSONB attribute search (e.g. filter by brand, category, rating)
CREATE INDEX idx_resources_attributes
    ON resources USING GIN (resource_attributes jsonb_path_ops);

CREATE INDEX idx_resources_catalog
    ON resources (catalog_id, bpp_id)
    WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- offers
-- ---------------------------------------------------------------------------

-- Array containment: find offers for a specific resource_id
CREATE INDEX idx_offers_resource_ids
    ON offers USING GIN (resource_ids);

CREATE INDEX idx_offers_catalog
    ON offers (catalog_id, bpp_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_offers_validity
    ON offers (validity_start, validity_end)
    WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- contracts (most read-heavy table in the BPP)
-- ---------------------------------------------------------------------------

-- Primary BAP lookup: find all contracts for a given BAP
CREATE INDEX idx_contracts_bap_id
    ON contracts (bap_id, status)
    WHERE deleted_at IS NULL;

-- Session correlation: all contracts in a user session
CREATE INDEX idx_contracts_transaction_id
    ON contracts (transaction_id);

-- Status-based querying (e.g. ACTIVE contracts for SLA monitoring)
CREATE INDEX idx_contracts_status_created
    ON contracts (status, created_at DESC)
    WHERE deleted_at IS NULL;

-- BPP-level contract listing
CREATE INDEX idx_contracts_bpp_status
    ON contracts (bpp_id, status, created_at DESC)
    WHERE deleted_at IS NULL;

-- Domain-specific filtering (e.g. all mobility contracts)
CREATE INDEX idx_contracts_domain
    ON contracts (domain, status)
    WHERE domain IS NOT NULL AND deleted_at IS NULL;

-- JSONB extension attribute search
CREATE INDEX idx_contracts_attributes
    ON contracts USING GIN (contract_attributes jsonb_path_ops)
    WHERE contract_attributes IS NOT NULL;

-- ---------------------------------------------------------------------------
-- commitments
-- ---------------------------------------------------------------------------

CREATE INDEX idx_commitments_contract
    ON commitments (contract_id, status);

-- ---------------------------------------------------------------------------
-- commitment_resources
-- ---------------------------------------------------------------------------

CREATE INDEX idx_commitment_resources_commitment
    ON commitment_resources (commitment_id, contract_id);

CREATE INDEX idx_commitment_resources_resource
    ON commitment_resources (resource_id);

-- ---------------------------------------------------------------------------
-- considerations
-- ---------------------------------------------------------------------------

CREATE INDEX idx_considerations_contract
    ON considerations (contract_id, status_code);

-- ---------------------------------------------------------------------------
-- contract_participants
-- ---------------------------------------------------------------------------

CREATE INDEX idx_contract_participants_contract
    ON contract_participants (contract_id);

-- ---------------------------------------------------------------------------
-- contract_performances
-- ---------------------------------------------------------------------------

CREATE INDEX idx_contract_performances_contract
    ON contract_performances (contract_id, status_code);

-- Array search: find performance by commitment_id
CREATE INDEX idx_contract_performances_commitment_ids
    ON contract_performances USING GIN (commitment_ids);

-- ---------------------------------------------------------------------------
-- settlements
-- ---------------------------------------------------------------------------

CREATE INDEX idx_settlements_contract
    ON settlements (contract_id, status);

CREATE INDEX idx_settlements_consideration
    ON settlements (consideration_id, contract_id);

-- ---------------------------------------------------------------------------
-- ratings
-- ---------------------------------------------------------------------------

CREATE INDEX idx_ratings_contract
    ON ratings (contract_id, created_at DESC);

CREATE INDEX idx_ratings_target
    ON ratings (target_id, created_at DESC);

CREATE INDEX idx_ratings_bap
    ON ratings (bap_id, created_at DESC);

-- ---------------------------------------------------------------------------
-- support_tickets
-- ---------------------------------------------------------------------------

CREATE INDEX idx_support_tickets_contract
    ON support_tickets (contract_id)
    WHERE contract_id IS NOT NULL;

CREATE INDEX idx_support_tickets_bap_status
    ON support_tickets (bap_id, status);

CREATE INDEX idx_support_tickets_status
    ON support_tickets (status, created_at DESC)
    WHERE status IN ('OPEN', 'IN_PROGRESS');

-- ---------------------------------------------------------------------------
-- contract_tracking
-- ---------------------------------------------------------------------------

CREATE INDEX idx_contract_tracking_contract
    ON contract_tracking (contract_id, status)
    WHERE status = 'ACTIVE';

-- ---------------------------------------------------------------------------
-- beckn_message_log
-- ---------------------------------------------------------------------------

-- transaction_id lookup (spans multiple partitions — local on each partition via inheritance)
CREATE INDEX idx_beckn_message_log_transaction_id
    ON beckn_message_log (transaction_id)
    WHERE transaction_id IS NOT NULL;

CREATE INDEX idx_beckn_message_log_message_id
    ON beckn_message_log (message_id);

CREATE INDEX idx_beckn_message_log_action_direction
    ON beckn_message_log (action, direction, created_at DESC);

-- BRIN index on created_at for time-range scans across the partitioned table
CREATE INDEX idx_beckn_message_log_created_at_brin
    ON beckn_message_log USING BRIN (created_at);

-- ---------------------------------------------------------------------------
-- outbox_events
-- ---------------------------------------------------------------------------

-- Worker polling: PENDING events ordered by next_attempt_at
CREATE INDEX idx_outbox_events_pending
    ON outbox_events (next_attempt_at ASC, status)
    WHERE status IN ('PENDING', 'FAILED');

CREATE INDEX idx_outbox_events_contract
    ON outbox_events (contract_id)
    WHERE contract_id IS NOT NULL;

-- ---------------------------------------------------------------------------
-- catalog_subscriptions
-- ---------------------------------------------------------------------------

-- Lookup subscriptions matching a given network during catalog publish
CREATE INDEX idx_catalog_subscriptions_network_ids
    ON catalog_subscriptions USING GIN (network_ids);

-- Schema type filter (GIN for array containment)
CREATE INDEX idx_catalog_subscriptions_schema_types
    ON catalog_subscriptions USING GIN (schema_types);

CREATE INDEX idx_catalog_subscriptions_subscriber
    ON catalog_subscriptions (subscriber_id, status);

CREATE INDEX idx_catalog_subscriptions_active
    ON catalog_subscriptions (status)
    WHERE status = 'ACTIVE';

---- create above / drop below ----

-- Drop all indexes (tables are dropped by their own migrations; indexes go with them)
-- Only need to drop indexes on tables that persist but lose their indexes
DROP INDEX IF EXISTS idx_bpp_identities_active;
DROP INDEX IF EXISTS idx_bpp_identities_network_ids;
DROP INDEX IF EXISTS idx_bpp_signing_keys_lookup;
DROP INDEX IF EXISTS idx_providers_bpp_id;
DROP INDEX IF EXISTS idx_providers_active;
DROP INDEX IF EXISTS idx_provider_locations_provider;
DROP INDEX IF EXISTS idx_provider_locations_geo;
DROP INDEX IF EXISTS idx_provider_locations_locality;
DROP INDEX IF EXISTS idx_provider_locations_country;
DROP INDEX IF EXISTS idx_catalogs_bpp_active;
DROP INDEX IF EXISTS idx_catalogs_provider;
DROP INDEX IF EXISTS idx_catalogs_validity;
DROP INDEX IF EXISTS idx_resources_search_vector;
DROP INDEX IF EXISTS idx_resources_attributes;
DROP INDEX IF EXISTS idx_resources_catalog;
DROP INDEX IF EXISTS idx_offers_resource_ids;
DROP INDEX IF EXISTS idx_offers_catalog;
DROP INDEX IF EXISTS idx_offers_validity;
DROP INDEX IF EXISTS idx_contracts_bap_id;
DROP INDEX IF EXISTS idx_contracts_transaction_id;
DROP INDEX IF EXISTS idx_contracts_status_created;
DROP INDEX IF EXISTS idx_contracts_bpp_status;
DROP INDEX IF EXISTS idx_contracts_domain;
DROP INDEX IF EXISTS idx_contracts_attributes;
DROP INDEX IF EXISTS idx_commitments_contract;
DROP INDEX IF EXISTS idx_commitment_resources_commitment;
DROP INDEX IF EXISTS idx_commitment_resources_resource;
DROP INDEX IF EXISTS idx_considerations_contract;
DROP INDEX IF EXISTS idx_contract_participants_contract;
DROP INDEX IF EXISTS idx_contract_performances_contract;
DROP INDEX IF EXISTS idx_contract_performances_commitment_ids;
DROP INDEX IF EXISTS idx_settlements_contract;
DROP INDEX IF EXISTS idx_settlements_consideration;
DROP INDEX IF EXISTS idx_ratings_contract;
DROP INDEX IF EXISTS idx_ratings_target;
DROP INDEX IF EXISTS idx_ratings_bap;
DROP INDEX IF EXISTS idx_support_tickets_contract;
DROP INDEX IF EXISTS idx_support_tickets_bap_status;
DROP INDEX IF EXISTS idx_support_tickets_status;
DROP INDEX IF EXISTS idx_contract_tracking_contract;
DROP INDEX IF EXISTS idx_beckn_message_log_transaction_id;
DROP INDEX IF EXISTS idx_beckn_message_log_message_id;
DROP INDEX IF EXISTS idx_beckn_message_log_action_direction;
DROP INDEX IF EXISTS idx_beckn_message_log_created_at_brin;
DROP INDEX IF EXISTS idx_outbox_events_pending;
DROP INDEX IF EXISTS idx_outbox_events_contract;
DROP INDEX IF EXISTS idx_catalog_subscriptions_network_ids;
DROP INDEX IF EXISTS idx_catalog_subscriptions_schema_types;
DROP INDEX IF EXISTS idx_catalog_subscriptions_subscriber;
DROP INDEX IF EXISTS idx_catalog_subscriptions_active;
