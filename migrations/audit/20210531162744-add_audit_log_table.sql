-- +migrate Up
CREATE EXTENSION IF NOT EXISTS pg_partman SCHEMA {{ .SCHEMA }};

CREATE TABLE _audit_log (
  -- Normally this should be PRIMARY KEY, but a partitioned table cannot have unique index on column that is not part of the partition key.
  id text NOT NULL,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  user_id text NOT NULL,
  activity_type text NOT NULL,
  ip_address inet,
  user_agent text,
  client_id text,
  data jsonb NOT NULL
) PARTITION BY RANGE (created_at);
CREATE INDEX _audit_log_idx_created_at_brin ON _audit_log USING BRIN (created_at);

CREATE TABLE _audit_log_template (LIKE _audit_log);
ALTER TABLE _audit_log_template ADD PRIMARY KEY (id);

SELECT create_parent(
  '{{ .SCHEMA }}._audit_log',
  'created_at',
  'native',
  'monthly',
  p_template_table := '{{ .SCHEMA }}._audit_log_template'
);

UPDATE part_config
SET retention = '90 days', retention_keep_table = FALSE
WHERE parent_table = '{{ .SCHEMA }}._audit_log';

-- +migrate Down
SELECT undo_partition(
  '{{ .SCHEMA }}._audit_log',
  p_keep_table := FALSE,
  p_target_table := '{{ .SCHEMA }}._audit_log_default'
);
-- Intuitively we should DROP _audit_log_template ourselves
-- because we created it ourselves.
-- However, undo_partition WILL drop it for us.
-- So if we drop it here again, the table does not exist.
-- DROP TABLE _audit_log_template;
DROP TABLE _audit_log;
