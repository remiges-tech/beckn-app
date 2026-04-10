-- Support tickets created via /support or /on_support.
-- A ticket may be linked to a contract (or raised standalone for a provider/resource).
-- channels JSONB stores the available support contact methods (email, phone, chat URL, etc.)
-- as an array of JSON-LD Attributes objects.

CREATE TABLE support_tickets (
    id                          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Optional contract linkage (support may be for a specific order)
    contract_id                 UUID        REFERENCES contracts (id),

    bap_id                      TEXT        NOT NULL,
    bap_uri                     TEXT        NOT NULL,

    -- Nature of the support request (Descriptor)
    descriptor_name             TEXT,
    descriptor_code             TEXT,
    descriptor_short_desc       TEXT,
    descriptor_long_desc        TEXT,

    -- Available support channels (array of JSON-LD Attributes objects)
    channels                    JSONB       NOT NULL DEFAULT '[]',

    status                      support_ticket_status NOT NULL DEFAULT 'OPEN',

    resolved_at                 TIMESTAMPTZ,

    -- Whether this was a preview (context.try: true) — preview returns channels without creating ticket
    is_preview                  BOOLEAN     NOT NULL DEFAULT FALSE,

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE support_tickets IS
    'Support tickets raised by BAPs against contracts or other entities. '
    'Channels carry JSON-LD contact descriptors (email, phone, chat, etc.).';

---- create above / drop below ----

DROP TABLE IF EXISTS support_tickets;
