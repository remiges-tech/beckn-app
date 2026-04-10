-- Catalog subscriptions: BAP subscriptions to catalog update events from this BPP.
-- A BAP subscribes by calling /catalog/subscription with networkIds and optional schemaTypes.
-- The BPP delivers on_discover callbacks for matching catalogs to the BAP's callbackUrl.
-- delivery_policy stores retry and timeout configuration (DeliveryPolicy schema).

CREATE TABLE catalog_subscriptions (
    id                  UUID                PRIMARY KEY DEFAULT gen_random_uuid(),

    -- BAP subscriber identity (from context.bapId / context.bapUri)
    subscriber_id       TEXT                NOT NULL,
    bap_id              TEXT                NOT NULL,
    bap_uri             TEXT                NOT NULL,

    -- Subscription filter criteria
    network_ids         TEXT[]              NOT NULL DEFAULT '{}',

    -- Schema type URIs; empty array = wildcard (match all types)
    schema_types        TEXT[]              NOT NULL DEFAULT '{}',

    -- Callback delivery target (defaults to bap_uri)
    callback_url        TEXT                NOT NULL,

    status              subscription_status NOT NULL DEFAULT 'ACTIVE',

    -- DeliveryPolicy (timeout, maxPayload, retryPolicy)
    delivery_policy     JSONB,

    created_at          TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ         NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE catalog_subscriptions IS
    'BAP subscriptions to catalog update events. '
    'Drives which on_discover callbacks are pushed when catalogs are published or updated.';

COMMENT ON COLUMN catalog_subscriptions.schema_types IS
    'JSON-LD schema type URIs for filtering. Empty array is treated as wildcard.';

COMMENT ON COLUMN catalog_subscriptions.delivery_policy IS
    'Retry and timeout policy per Beckn DeliveryPolicy schema.';

---- create above / drop below ----

DROP TABLE IF EXISTS catalog_subscriptions;
