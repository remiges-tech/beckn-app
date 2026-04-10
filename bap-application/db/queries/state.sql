-- Queries for persisting contract state received in on_* callbacks.

-- name: UpsertTransaction :exec
-- Creates or updates the transaction tracking row.
INSERT INTO transactions (transaction_id, bap_id, network_id, bpp_id, bpp_uri, status)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (transaction_id) DO UPDATE SET
    bpp_id     = EXCLUDED.bpp_id,
    bpp_uri    = EXCLUDED.bpp_uri,
    status     = EXCLUDED.status,
    updated_at = NOW();

-- name: UpsertContractSnapshot :exec
-- Stores (or replaces) the full contract JSON for a given (transaction_id, action).
INSERT INTO contract_snapshots (transaction_id, action, contract_id, contract)
VALUES ($1, $2, $3, $4)
ON CONFLICT (transaction_id, action) DO UPDATE SET
    contract_id = EXCLUDED.contract_id,
    contract    = EXCLUDED.contract,
    updated_at  = NOW();

-- name: GetTransaction :one
-- Returns the transaction row for the given transaction UUID.
SELECT transaction_id, bap_id, network_id, bpp_id, bpp_uri, status, created_at, updated_at
FROM transactions
WHERE transaction_id = $1;

-- name: GetContractSnapshot :one
-- Returns the most recent contract snapshot for a (transaction_id, action) pair.
SELECT id, transaction_id, action, contract_id, contract, created_at, updated_at
FROM contract_snapshots
WHERE transaction_id = $1
  AND action = $2;

-- name: GetLatestContractSnapshot :one
-- Returns the latest snapshot for a transaction (highest priority action).
SELECT id, transaction_id, action, contract_id, contract, created_at, updated_at
FROM contract_snapshots
WHERE transaction_id = $1
ORDER BY created_at DESC
LIMIT 1;
