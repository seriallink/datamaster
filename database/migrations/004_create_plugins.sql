-- Creates required PostgreSQL extensions for replication and observability

CREATE EXTENSION IF NOT EXISTS pglogical;
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;