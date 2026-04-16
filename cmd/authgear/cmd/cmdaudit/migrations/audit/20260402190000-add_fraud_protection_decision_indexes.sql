-- +migrate Up
CREATE INDEX _audit_log_idx_fraud_decision_app_id_decision_created_at
  ON _audit_log (
    app_id,
    (data #>> '{payload,record,decision}'),
    created_at DESC
  )
  WHERE activity_type = 'fraud_protection.decision_recorded';

CREATE INDEX _audit_log_idx_fraud_decision_app_id_action_created_at
  ON _audit_log (
    app_id,
    (data #>> '{payload,record,action}'),
    created_at DESC
  )
  WHERE activity_type = 'fraud_protection.decision_recorded';

CREATE INDEX _audit_log_idx_fraud_decision_app_id_recipient_created_at
  ON _audit_log (
    app_id,
    (data #>> '{payload,record,action_detail,recipient}') text_pattern_ops,
    created_at DESC
  )
  WHERE activity_type = 'fraud_protection.decision_recorded';

CREATE INDEX _audit_log_idx_fraud_decision_app_id_phone_country_created_at
  ON _audit_log (
    app_id,
    (upper(data #>> '{payload,record,action_detail,phone_number_country_code}')) text_pattern_ops,
    created_at DESC
  )
  WHERE activity_type = 'fraud_protection.decision_recorded';

CREATE INDEX _audit_log_idx_fraud_decision_app_id_geo_country_created_at
  ON _audit_log (
    app_id,
    (upper(data #>> '{payload,record,geo_location_code}')) text_pattern_ops,
    created_at DESC
  )
  WHERE activity_type = 'fraud_protection.decision_recorded';

CREATE INDEX _audit_log_idx_fraud_decision_app_id_ip_created_at
  ON _audit_log (
    app_id,
    (ip_address::text) text_pattern_ops,
    created_at DESC
  )
  WHERE activity_type = 'fraud_protection.decision_recorded';

-- +migrate Down
DROP INDEX _audit_log_idx_fraud_decision_app_id_ip_created_at;

DROP INDEX _audit_log_idx_fraud_decision_app_id_geo_country_created_at;

DROP INDEX _audit_log_idx_fraud_decision_app_id_phone_country_created_at;

DROP INDEX _audit_log_idx_fraud_decision_app_id_recipient_created_at;

DROP INDEX _audit_log_idx_fraud_decision_app_id_action_created_at;

DROP INDEX _audit_log_idx_fraud_decision_app_id_decision_created_at;
