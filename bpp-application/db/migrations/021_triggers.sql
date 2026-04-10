-- Triggers for automated housekeeping:
-- 1. updated_at maintenance — every mutable table gets an auto-refresh trigger.
-- 2. resources search_vector — weighted tsvector from name, short/long desc, and attributes.

-- ---------------------------------------------------------------------------
-- Helper function: touch updated_at on any row change
-- ---------------------------------------------------------------------------

CREATE OR REPLACE FUNCTION fn_set_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$;

-- ---------------------------------------------------------------------------
-- updated_at triggers (all mutable tables)
-- ---------------------------------------------------------------------------

CREATE TRIGGER trg_bpp_identities_updated_at
    BEFORE UPDATE ON bpp_identities
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

CREATE TRIGGER trg_bpp_signing_keys_updated_at
    BEFORE UPDATE ON bpp_signing_keys
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

CREATE TRIGGER trg_providers_updated_at
    BEFORE UPDATE ON providers
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

CREATE TRIGGER trg_provider_locations_updated_at
    BEFORE UPDATE ON provider_locations
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

CREATE TRIGGER trg_catalogs_updated_at
    BEFORE UPDATE ON catalogs
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

CREATE TRIGGER trg_resources_updated_at
    BEFORE UPDATE ON resources
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

CREATE TRIGGER trg_offers_updated_at
    BEFORE UPDATE ON offers
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

CREATE TRIGGER trg_contracts_updated_at
    BEFORE UPDATE ON contracts
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

CREATE TRIGGER trg_commitments_updated_at
    BEFORE UPDATE ON commitments
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

CREATE TRIGGER trg_commitment_resources_updated_at
    BEFORE UPDATE ON commitment_resources
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

CREATE TRIGGER trg_considerations_updated_at
    BEFORE UPDATE ON considerations
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

CREATE TRIGGER trg_contract_participants_updated_at
    BEFORE UPDATE ON contract_participants
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

CREATE TRIGGER trg_contract_performances_updated_at
    BEFORE UPDATE ON contract_performances
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

CREATE TRIGGER trg_settlements_updated_at
    BEFORE UPDATE ON settlements
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

CREATE TRIGGER trg_support_tickets_updated_at
    BEFORE UPDATE ON support_tickets
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

CREATE TRIGGER trg_contract_tracking_updated_at
    BEFORE UPDATE ON contract_tracking
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

CREATE TRIGGER trg_outbox_events_updated_at
    BEFORE UPDATE ON outbox_events
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

CREATE TRIGGER trg_catalog_subscriptions_updated_at
    BEFORE UPDATE ON catalog_subscriptions
    FOR EACH ROW EXECUTE FUNCTION fn_set_updated_at();

-- ---------------------------------------------------------------------------
-- resources.search_vector — weighted full-text search
--
-- Weight mapping:
--   A (highest) — descriptor_name
--   B           — descriptor_short_desc
--   C           — descriptor_long_desc
--   D (lowest)  — descriptor_code + resource_attributes_type
-- ---------------------------------------------------------------------------

CREATE OR REPLACE FUNCTION fn_resources_search_vector_update()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', coalesce(NEW.descriptor_name, '')),     'A') ||
        setweight(to_tsvector('english', coalesce(NEW.descriptor_short_desc, '')), 'B') ||
        setweight(to_tsvector('english', coalesce(NEW.descriptor_long_desc, '')),  'C') ||
        setweight(to_tsvector('english', coalesce(NEW.descriptor_code, '')),        'D') ||
        setweight(to_tsvector('english', coalesce(NEW.resource_attributes_type, '')), 'D');
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_resources_search_vector
    BEFORE INSERT OR UPDATE
    OF descriptor_name, descriptor_short_desc, descriptor_long_desc,
       descriptor_code, resource_attributes_type
    ON resources
    FOR EACH ROW EXECUTE FUNCTION fn_resources_search_vector_update();

---- create above / drop below ----

-- Drop triggers in reverse order
DROP TRIGGER IF EXISTS trg_resources_search_vector          ON resources;
DROP TRIGGER IF EXISTS trg_catalog_subscriptions_updated_at ON catalog_subscriptions;
DROP TRIGGER IF EXISTS trg_outbox_events_updated_at         ON outbox_events;
DROP TRIGGER IF EXISTS trg_contract_tracking_updated_at     ON contract_tracking;
DROP TRIGGER IF EXISTS trg_support_tickets_updated_at       ON support_tickets;
DROP TRIGGER IF EXISTS trg_settlements_updated_at           ON settlements;
DROP TRIGGER IF EXISTS trg_contract_performances_updated_at ON contract_performances;
DROP TRIGGER IF EXISTS trg_contract_participants_updated_at ON contract_participants;
DROP TRIGGER IF EXISTS trg_considerations_updated_at        ON considerations;
DROP TRIGGER IF EXISTS trg_commitment_resources_updated_at  ON commitment_resources;
DROP TRIGGER IF EXISTS trg_commitments_updated_at           ON commitments;
DROP TRIGGER IF EXISTS trg_contracts_updated_at             ON contracts;
DROP TRIGGER IF EXISTS trg_offers_updated_at                ON offers;
DROP TRIGGER IF EXISTS trg_resources_updated_at             ON resources;
DROP TRIGGER IF EXISTS trg_catalogs_updated_at              ON catalogs;
DROP TRIGGER IF EXISTS trg_provider_locations_updated_at    ON provider_locations;
DROP TRIGGER IF EXISTS trg_providers_updated_at             ON providers;
DROP TRIGGER IF EXISTS trg_bpp_signing_keys_updated_at      ON bpp_signing_keys;
DROP TRIGGER IF EXISTS trg_bpp_identities_updated_at        ON bpp_identities;

DROP FUNCTION IF EXISTS fn_resources_search_vector_update();
DROP FUNCTION IF EXISTS fn_set_updated_at();
