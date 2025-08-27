-- +migrate Up
UPDATE part_config
SET retention = '180 days'
WHERE parent_table = '{{ .SCHEMA }}._audit_log';

-- +migrate Down
UPDATE part_config
SET retention = '90 days'
WHERE parent_table = '{{ .SCHEMA }}._audit_log';
