-- Queries for the /beckn/confirm handler.
-- Confirm transitions the contract from DRAFT → ACTIVE.
-- The contract MUST already exist (created during init).

-- name: ConfirmContract :one
-- Transitions a DRAFT contract to ACTIVE status.
-- Returns the contract id on success; errors if no DRAFT contract is found.
UPDATE contracts SET
    status     = 'ACTIVE'::contract_status,
    updated_at = NOW()
WHERE transaction_id = $1
  AND bpp_id = $2
  AND status = 'DRAFT'::contract_status
  AND deleted_at IS NULL
RETURNING id;
