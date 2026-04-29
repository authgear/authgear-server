-- +migrate Up

ALTER TABLE _portal_app_collaborator ADD COLUMN updated_at timestamp with time zone;
UPDATE _portal_app_collaborator SET updated_at = created_at;
ALTER TABLE _portal_app_collaborator ALTER COLUMN updated_at SET NOT NULL;

-- +migrate Down

ALTER TABLE _portal_app_collaborator DROP COLUMN updated_at;
