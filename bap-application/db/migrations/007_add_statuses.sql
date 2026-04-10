-- Add new transaction lifecycle statuses for post-confirm actions.

ALTER TYPE transaction_status ADD VALUE IF NOT EXISTS 'STATUS_SENT';
ALTER TYPE transaction_status ADD VALUE IF NOT EXISTS 'CANCEL_SENT';
ALTER TYPE transaction_status ADD VALUE IF NOT EXISTS 'UPDATE_SENT';
ALTER TYPE transaction_status ADD VALUE IF NOT EXISTS 'RATE_SENT';
ALTER TYPE transaction_status ADD VALUE IF NOT EXISTS 'SUPPORT_SENT';

---- create above / drop below ----

-- Enum values cannot be dropped in PostgreSQL; migration is intentionally irreversible.
