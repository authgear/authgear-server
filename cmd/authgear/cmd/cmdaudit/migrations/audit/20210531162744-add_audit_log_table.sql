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

-- In case you wonder why we have to wrap this in StatementBegin and StatementEnd,
-- please read https://github.com/rubenv/sql-migrate/issues/261#issuecomment-1918115180
-- +migrate StatementBegin
DO LANGUAGE 'plpgsql' $$DECLARE
  pg_partman_version text;
BEGIN
  SELECT extversion INTO pg_partman_version FROM pg_extension WHERE extname = 'pg_partman';
  IF pg_partman_version like '5.%' THEN
    PERFORM create_parent(
      p_parent_table := '{{ .SCHEMA }}._audit_log',
      p_control := 'created_at',
      p_interval := '1 month',
      p_template_table := '{{ .SCHEMA }}._audit_log_template'
    );
    UPDATE part_config
    SET retention = '90 days', retention_keep_table = FALSE
    WHERE parent_table = '{{ .SCHEMA }}._audit_log';
  ELSIF pg_partman_version like '4.%' THEN
    PERFORM create_parent(
      '{{ .SCHEMA }}._audit_log',
      'created_at',
      'native',
      'monthly',
      p_template_table := '{{ .SCHEMA }}._audit_log_template'
    );

    UPDATE part_config
    SET retention = '90 days', retention_keep_table = FALSE
    WHERE parent_table = '{{ .SCHEMA }}._audit_log';
  ELSE
    RAISE EXCEPTION 'unsupported pg_partman version %', pg_partman_version;
  END IF;
END$$;
-- +migrate StatementEnd

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
