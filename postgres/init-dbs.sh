#!/usr/bin/env bash
# postgres/init-dbs.sh
# Creates the bpp and bap databases inside the shared Postgres container.
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    SELECT 'CREATE DATABASE bpp' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'bpp')\gexec
    SELECT 'CREATE DATABASE bap' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'bap')\gexec
EOSQL
