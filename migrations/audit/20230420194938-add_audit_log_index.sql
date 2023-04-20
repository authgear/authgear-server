-- +migrate Up
CREATE INDEX _audit_log_idx_app_id_activity_type_created_at
  ON _audit_log (app_id, activity_type, created_at);

-- +migrate Down
DROP INDEX _audit_log_idx_app_id_activity_type_created_at_brin;
