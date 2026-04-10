-- BAP infrastructure queries: identity seeding and message audit logging.

-- name: UpsertBAPIdentity :exec
INSERT INTO bap_identities (bap_id, bap_uri, network_ids, is_active)
VALUES ($1, $2, $3, true)
ON CONFLICT (bap_id) DO UPDATE SET
    bap_uri     = EXCLUDED.bap_uri,
    network_ids = EXCLUDED.network_ids,
    updated_at  = NOW();

-- name: InsertMessageLog :exec
INSERT INTO beckn_message_log (
    transaction_id, message_id,
    action, direction,
    url,
    bap_id, bap_uri,
    bpp_id, bpp_uri, network_id,
    ack_status,
    request_payload, response_payload,
    error_message,
    processing_duration_ms
) VALUES (
    $1, $2,
    $3, $4,
    $5,
    $6, $7,
    $8, $9, $10,
    $11,
    $12, $13,
    $14,
    $15
);
