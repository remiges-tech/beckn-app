-- Enable required PostgreSQL extensions for the BPP application.

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

---- create above / drop below ----

DROP EXTENSION IF EXISTS "uuid-ossp";
