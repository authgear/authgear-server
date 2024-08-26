-- +migrate Up
CREATE INDEX _audit_log_idx_app_id_activity_type_data_payload_recipient_created_at ON _audit_log (
  app_id,
  activity_type,
  (data #>> '{payload,recipient}'),
  created_at DESC
);

-- +migrate Down
DROP INDEX _audit_log_idx_app_id_activity_type_data_payload_recipient_created_at;
