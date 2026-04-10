-- Add url column to beckn_message_log.
-- Stores the endpoint URL for each Beckn protocol call:
--   INBOUND  → the path on this BPP that was called (e.g. /v1/catalog/publish)
--   OUTBOUND → the full URL of the external service called (e.g. the CDS publish URL)

ALTER TABLE beckn_message_log ADD COLUMN IF NOT EXISTS url TEXT;

---- create above / drop below ----

ALTER TABLE beckn_message_log DROP COLUMN IF EXISTS url;
