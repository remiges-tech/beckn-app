-- BPP infrastructure queries: identity seeding and message audit logging.

-- name: UpsertBPPIdentity :exec
-- Ensures this BPP's subscriber record exists before any tenant data is written.
-- Called once at application startup using values from config / environment variables.
INSERT INTO bpp_identities (bpp_id, bpp_uri, network_ids, is_active)
VALUES ($1, $2, $3, true)
ON CONFLICT (bpp_id) DO UPDATE SET
    bpp_uri     = EXCLUDED.bpp_uri,
    network_ids = EXCLUDED.network_ids,
    updated_at  = NOW();

-- name: InsertMessageLog :exec
-- Appends one immutable row to the beckn_message_log audit table.
-- Called only for Beckn protocol API calls:
--   INBOUND  — a Beckn participant called one of our hosted endpoints
--   OUTBOUND — we called an external Beckn endpoint (CDS, BAP callback, etc.)
-- Internal-only operations (DB queries, cache, etc.) are NOT logged here.
-- Failures are warnings and must never abort the main request flow.
INSERT INTO beckn_message_log (
    transaction_id, message_id,
    action, direction,
    url,
    bap_id, bap_uri,
    bpp_id, bpp_uri, network_id,
    ack_status,
    request_payload, response_payload,
    error_code, error_message,
    processing_duration_ms
) VALUES (
    $1, $2,
    $3, $4,
    $5,
    $6, $7,
    $8, $9, $10,
    $11,
    $12, $13,
    $14, $15,
    $16
);
