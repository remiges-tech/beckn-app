-- Queries for the /beckn/init handler.
-- These persist contract state derived from the BAP's init request.
-- The transactionId (from Beckn context) is the stable key across the
-- discover→select→init→confirm lifecycle.

-- name: GetContractByTransactionID :one
-- Look up the existing DRAFT contract for this session (if any).
SELECT id FROM contracts
WHERE transaction_id = $1
  AND bpp_id = $2
  AND status = 'DRAFT'::contract_status
  AND deleted_at IS NULL
LIMIT 1;

-- name: InsertContract :one
-- Create a new contract row for this init request.
INSERT INTO contracts (
    id,
    bpp_id,
    bap_id,
    bap_uri,
    transaction_id,
    network_id,
    status,
    contract_attributes,
    contract_attributes_context,
    contract_attributes_type
)
VALUES (
    $1, $2, $3, $4, $5, $6,
    'DRAFT'::contract_status,
    $7, $8, $9
)
RETURNING id;

-- name: UpdateContractAttributes :exec
-- Refresh mutable fields on an existing contract (called when contract already exists).
UPDATE contracts SET
    bap_uri                     = $2,
    contract_attributes         = $3,
    contract_attributes_context = $4,
    contract_attributes_type    = $5,
    updated_at                  = NOW()
WHERE id = $1;

-- name: UpsertParticipant :exec
-- Inserts or refreshes a participant row for the given contract.
INSERT INTO contract_participants (
    id,
    contract_id,
    descriptor_name,
    participant_attributes,
    participant_attributes_context,
    participant_attributes_type
)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id, contract_id)
    DO UPDATE SET
        descriptor_name                = EXCLUDED.descriptor_name,
        participant_attributes         = EXCLUDED.participant_attributes,
        participant_attributes_context = EXCLUDED.participant_attributes_context,
        participant_attributes_type    = EXCLUDED.participant_attributes_type,
        updated_at                     = NOW();

-- name: UpsertCommitment :exec
-- Inserts or refreshes a commitment row linked to the contract.
INSERT INTO commitments (
    id,
    contract_id,
    status,
    commitment_attributes,
    commitment_attributes_context,
    commitment_attributes_type
)
VALUES (
    $1, $2,
    'DRAFT'::commitment_status,
    $3, $4, $5
)
ON CONFLICT (id, contract_id)
    DO UPDATE SET
        commitment_attributes         = EXCLUDED.commitment_attributes,
        commitment_attributes_context = EXCLUDED.commitment_attributes_context,
        commitment_attributes_type    = EXCLUDED.commitment_attributes_type,
        updated_at                    = NOW();

-- name: UpsertCommitmentOffer :exec
-- Inserts or refreshes the offer snapshot for a commitment.
INSERT INTO commitment_offers (
    commitment_id,
    contract_id,
    offer_id,
    offer_descriptor,
    offer_attributes,
    resource_ids
)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (commitment_id, contract_id)
    DO UPDATE SET
        offer_id         = EXCLUDED.offer_id,
        offer_descriptor = EXCLUDED.offer_descriptor,
        offer_attributes = EXCLUDED.offer_attributes,
        resource_ids     = EXCLUDED.resource_ids;

-- name: UpsertConsideration :exec
-- Inserts or refreshes a consideration (pricing) row for the contract.
INSERT INTO considerations (
    id,
    contract_id,
    status_code,
    consideration_attributes,
    consideration_attributes_context,
    consideration_attributes_type
)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id, contract_id)
    DO UPDATE SET
        status_code                      = EXCLUDED.status_code,
        consideration_attributes         = EXCLUDED.consideration_attributes,
        consideration_attributes_context = EXCLUDED.consideration_attributes_context,
        consideration_attributes_type    = EXCLUDED.consideration_attributes_type,
        updated_at                       = NOW();

-- name: UpsertPerformance :exec
-- Inserts or refreshes a performance (fulfillment) unit for the contract.
INSERT INTO contract_performances (
    id,
    contract_id,
    status_code,
    commitment_ids,
    performance_attributes,
    performance_attributes_context,
    performance_attributes_type
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (id, contract_id)
    DO UPDATE SET
        status_code                     = EXCLUDED.status_code,
        commitment_ids                  = EXCLUDED.commitment_ids,
        performance_attributes          = EXCLUDED.performance_attributes,
        performance_attributes_context  = EXCLUDED.performance_attributes_context,
        performance_attributes_type     = EXCLUDED.performance_attributes_type,
        updated_at                      = NOW();
