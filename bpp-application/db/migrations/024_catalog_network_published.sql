-- Track when a catalog was last forwarded to the CDS (network publish).
-- NULL means never published; a timestamp means it was successfully published.
ALTER TABLE catalogs ADD COLUMN IF NOT EXISTS network_published_at TIMESTAMPTZ;
