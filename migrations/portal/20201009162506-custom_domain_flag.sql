-- +migrate Up

ALTER TABLE _portal_pending_domain ADD COLUMN is_custom boolean;
UPDATE _portal_pending_domain SET is_custom = false;
ALTER TABLE _portal_pending_domain ALTER COLUMN is_custom SET NOT NULL;

ALTER TABLE _portal_domain ADD COLUMN is_custom boolean;
UPDATE _portal_domain SET is_custom = false;
ALTER TABLE _portal_domain ALTER COLUMN is_custom SET NOT NULL;

-- +migrate Down

ALTER TABLE _portal_domain DROP COLUMN is_custom;
ALTER TABLE _portal_pending_domain DROP COLUMN is_custom;
