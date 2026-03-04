-- +migrate Up
CREATE TABLE _audit_metrics (
    -- Normally this should be PRIMARY KEY, but a partitioned table cannot have
    -- a unique index on a column that is not part of the partition key.
    id          TEXT                        NOT NULL,
    app_id      TEXT                        NOT NULL,
    name        TEXT                        NOT NULL,
    key         TEXT                        NOT NULL,
    created_at  TIMESTAMP WITHOUT TIME ZONE NOT NULL
) PARTITION BY RANGE (created_at);

CREATE INDEX _audit_metrics_idx ON _audit_metrics (app_id, name, key, created_at);

CREATE TABLE _audit_metrics_template (LIKE _audit_metrics);
ALTER TABLE _audit_metrics_template ADD PRIMARY KEY (id);

-- In case you wonder why we have to wrap this in StatementBegin and StatementEnd,
-- please read https://github.com/rubenv/sql-migrate/issues/261#issuecomment-1918115180
-- +migrate StatementBegin
DO LANGUAGE 'plpgsql' $$DECLARE
  pg_partman_version text;
BEGIN
  SELECT extversion INTO pg_partman_version FROM pg_extension WHERE extname = 'pg_partman';
  IF pg_partman_version like '5.%' THEN
    PERFORM create_parent(
      p_parent_table := '{{ .SCHEMA }}._audit_metrics',
      p_control := 'created_at',
      p_interval := '1 month',
      p_template_table := '{{ .SCHEMA }}._audit_metrics_template'
    );
    UPDATE part_config SET retention = '90 days', retention_keep_table = FALSE
    WHERE parent_table = '{{ .SCHEMA }}._audit_metrics';
  ELSIF pg_partman_version like '4.%' THEN
    PERFORM create_parent(
      '{{ .SCHEMA }}._audit_metrics', 'created_at', 'native', 'monthly',
      p_template_table := '{{ .SCHEMA }}._audit_metrics_template'
    );
    UPDATE part_config SET retention = '90 days', retention_keep_table = FALSE
    WHERE parent_table = '{{ .SCHEMA }}._audit_metrics';
  ELSE
    RAISE EXCEPTION 'unsupported pg_partman version %', pg_partman_version;
  END IF;
END$$;
-- +migrate StatementEnd

-- +migrate Down
SELECT undo_partition('{{ .SCHEMA }}._audit_metrics', p_keep_table := FALSE,
  p_target_table := '{{ .SCHEMA }}._audit_metrics_default');
DROP TABLE _audit_metrics;
