-- +migrate Up
ALTER TABLE _portal_subscription ADD COLUMN pending_update_since timestamp without time zone;

-- +migrate Down
ALTER TABLE _portal_subscription DROP COLUMN pending_update_since;
