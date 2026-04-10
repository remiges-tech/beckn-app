-- Stock management queries for resource_stock table.

-- name: UpsertResourceStock :exec
-- Called during catalog publish to set initial (or updated) stock quantity.
INSERT INTO resource_stock (resource_id, bpp_id, quantity, sold, updated_at)
VALUES ($1, $2, $3, 0, NOW())
ON CONFLICT (resource_id, bpp_id)
DO UPDATE SET
    quantity   = EXCLUDED.quantity,
    updated_at = NOW();

-- name: DecrementResourceStock :exec
-- Called on successful confirm to move one unit from available to sold.
-- Uses GREATEST(0, …) to prevent the counter going negative if called twice.
UPDATE resource_stock
SET quantity   = GREATEST(0, quantity - 1),
    sold       = sold + 1,
    updated_at = NOW()
WHERE resource_id = $1
  AND bpp_id      = $2;

-- name: ListResourceStock :many
-- Full stock listing for the dashboard, joined with the resource descriptor.
SELECT
    rs.resource_id,
    rs.bpp_id,
    rs.quantity,
    rs.sold,
    rs.updated_at,
    r.descriptor_name,
    r.catalog_id
FROM resource_stock rs
LEFT JOIN resources r
       ON r.id     = rs.resource_id
      AND r.bpp_id = rs.bpp_id
WHERE rs.bpp_id = $1
ORDER BY rs.updated_at DESC;
