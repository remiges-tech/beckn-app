-- Queries for fulfillment actions: status, cancel, update, rate, support.

-- ============================================================
-- STATUS — look up full contract for on_status callback
-- ============================================================

-- name: GetActiveContractByTxnID :one
-- Returns core contract fields needed to build an on_status payload.
SELECT
    c.id,
    c.bap_id,
    c.bap_uri,
    c.status,
    c.transaction_id,
    c.network_id,
    c.domain,
    c.contract_attributes,
    c.created_at,
    c.updated_at
FROM contracts c
WHERE c.transaction_id = $1
  AND c.bpp_id = $2
  AND c.deleted_at IS NULL
LIMIT 1;

-- ============================================================
-- CANCEL — transition ACTIVE → CANCELLED
-- ============================================================

-- name: CancelContract :one
-- Transitions an ACTIVE contract to CANCELLED.
-- Returns bap_uri and bap_id so the service can call on_cancel.
UPDATE contracts SET
    status     = 'CANCELLED'::contract_status,
    updated_at = NOW()
WHERE transaction_id = $1
  AND bpp_id = $2
  AND status = 'ACTIVE'::contract_status
  AND deleted_at IS NULL
RETURNING id, bap_id, bap_uri;

-- name: GetCommitmentResourceIDsByContractID :many
-- Returns all committed resource IDs for a contract (used to restore stock on cancel).
SELECT DISTINCT cr.resource_id
FROM commitment_resources cr
JOIN commitments cm ON cm.id = cr.commitment_id AND cm.contract_id = cr.contract_id
WHERE cm.contract_id = $1;

-- ============================================================
-- RATE — persist rating inputs
-- ============================================================

-- name: InsertRating :exec
INSERT INTO ratings (
    contract_id,
    bap_id,
    bap_uri,
    target_id,
    target_descriptor,
    range,
    feedback_form_submission,
    is_preview
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: ListRatings :many
SELECT
    r.id,
    r.contract_id,
    c.transaction_id,
    r.bap_id,
    r.target_id,
    r.target_descriptor,
    r.range,
    r.feedback_form_submission,
    r.is_preview,
    r.created_at
FROM ratings r
JOIN contracts c ON c.id = r.contract_id
WHERE c.bpp_id = $1
ORDER BY r.created_at DESC
LIMIT  $2
OFFSET $3;

-- name: CountRatings :one
SELECT COUNT(*) FROM ratings r
JOIN contracts c ON c.id = r.contract_id
WHERE c.bpp_id = $1;

-- ============================================================
-- SUPPORT — persist support tickets
-- ============================================================

-- name: InsertSupportTicket :one
INSERT INTO support_tickets (
    contract_id,
    bap_id,
    bap_uri,
    descriptor_name,
    descriptor_short_desc,
    channels,
    is_preview
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id;

-- name: ListSupportTickets :many
SELECT
    st.id,
    st.contract_id,
    c.transaction_id,
    st.bap_id,
    st.descriptor_name,
    st.descriptor_short_desc,
    st.channels,
    st.status,
    st.is_preview,
    st.created_at,
    st.resolved_at
FROM support_tickets st
LEFT JOIN contracts c ON c.id = st.contract_id
WHERE (c.bpp_id = $1 OR st.contract_id IS NULL)
ORDER BY st.created_at DESC
LIMIT  $2
OFFSET $3;

-- name: CountSupportTickets :one
SELECT COUNT(*) FROM support_tickets st
LEFT JOIN contracts c ON c.id = st.contract_id
WHERE c.bpp_id = $1 OR st.contract_id IS NULL;
