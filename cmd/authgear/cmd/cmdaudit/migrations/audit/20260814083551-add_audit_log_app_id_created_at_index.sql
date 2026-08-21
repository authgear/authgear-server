-- +migrate Up
CREATE INDEX _audit_log_idx_app_id_created_at ON _audit_log (app_id, created_at DESC);

-- +migrate Down
DROP INDEX _audit_log_idx_app_id_created_at;
