-- Domain enums derived directly from Beckn Protocol v2.0 specification.
-- Using native PG enums for type safety and storage efficiency.
-- Adding new values requires ALTER TYPE ... ADD VALUE — avoid renaming existing values.

-- Contract lifecycle states per spec §Contract.status
CREATE TYPE contract_status AS ENUM (
    'DRAFT',
    'ACTIVE',
    'CANCELLED',
    'COMPLETE'
);

-- Commitment states per spec §Commitment.status
CREATE TYPE commitment_status AS ENUM (
    'DRAFT',
    'ACTIVE',
    'CLOSED'
);

-- Settlement states per spec §Settlement.status
CREATE TYPE settlement_status AS ENUM (
    'DRAFT',
    'COMMITTED',
    'COMPLETE'
);

-- All Beckn protocol actions that can appear in Context.action
CREATE TYPE beckn_action AS ENUM (
    'discover',
    'on_discover',
    'select',
    'on_select',
    'init',
    'on_init',
    'confirm',
    'on_confirm',
    'status',
    'on_status',
    'track',
    'on_track',
    'update',
    'on_update',
    'cancel',
    'on_cancel',
    'rate',
    'on_rate',
    'support',
    'on_support',
    'catalog_publish',
    'catalog_on_publish',
    'catalog_subscription',
    'catalog_pull'
);

-- Synchronous response acknowledgement type
CREATE TYPE ack_status AS ENUM (
    'ACK',
    'NACK'
);

-- Direction of a Beckn message relative to this BPP
CREATE TYPE message_direction AS ENUM (
    'INBOUND',
    'OUTBOUND'
);

-- Catalog publish type per publishDirectives
CREATE TYPE catalog_type AS ENUM (
    'master',
    'regular'
);

-- Catalog update mode for regular overlays
CREATE TYPE catalog_update_mode AS ENUM (
    'FULL',
    'MERGE'
);

-- Catalog subscription lifecycle status
CREATE TYPE subscription_status AS ENUM (
    'ACTIVE',
    'INACTIVE',
    'DELETED'
);

-- Real-time tracking session status
CREATE TYPE tracking_status AS ENUM (
    'ACTIVE',
    'INACTIVE'
);

-- Outbox event delivery status
CREATE TYPE outbox_status AS ENUM (
    'PENDING',
    'PROCESSING',
    'DELIVERED',
    'FAILED',
    'DEAD_LETTERED'
);

-- Support ticket lifecycle
CREATE TYPE support_ticket_status AS ENUM (
    'OPEN',
    'IN_PROGRESS',
    'RESOLVED',
    'CLOSED'
);

---- create above / drop below ----

DROP TYPE IF EXISTS support_ticket_status;
DROP TYPE IF EXISTS outbox_status;
DROP TYPE IF EXISTS tracking_status;
DROP TYPE IF EXISTS subscription_status;
DROP TYPE IF EXISTS catalog_update_mode;
DROP TYPE IF EXISTS catalog_type;
DROP TYPE IF EXISTS message_direction;
DROP TYPE IF EXISTS ack_status;
DROP TYPE IF EXISTS beckn_action;
DROP TYPE IF EXISTS settlement_status;
DROP TYPE IF EXISTS commitment_status;
DROP TYPE IF EXISTS contract_status;
