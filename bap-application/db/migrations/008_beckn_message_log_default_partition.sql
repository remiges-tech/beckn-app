-- Safety-net partition: catches any row whose created_at falls outside the
-- explicit monthly partitions (e.g. no maintenance job has created next
-- month's partition yet). Without this, inserts outside existing ranges
-- fail with "no partition of relation found for row" (SQLSTATE 23514).

CREATE TABLE IF NOT EXISTS beckn_message_log_default
    PARTITION OF beckn_message_log DEFAULT;

---- create above / drop below ----

DROP TABLE IF EXISTS beckn_message_log_default;
