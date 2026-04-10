-- Inventory stock tracking per resource per BPP.
-- quantity = current on-hand stock (decremented on every confirmed order).
-- sold     = cumulative units sold (incremented on every confirmed order).
-- Stock is keyed by (resource_id, bpp_id) — a single physical SKU counter
-- shared across all catalogs that reference the same resource ID.

CREATE TABLE resource_stock (
    resource_id   TEXT        NOT NULL,
    bpp_id        TEXT        NOT NULL,
    quantity      INTEGER     NOT NULL DEFAULT 0,
    sold          INTEGER     NOT NULL DEFAULT 0,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (resource_id, bpp_id)
);

COMMENT ON TABLE resource_stock IS
    'Per-resource inventory counter. quantity = remaining stock; sold = cumulative confirmed units.';

---- create above / drop below ----

DROP TABLE IF EXISTS resource_stock;
