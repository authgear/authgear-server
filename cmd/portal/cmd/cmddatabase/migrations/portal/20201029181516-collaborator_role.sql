-- +migrate Up

ALTER TABLE _portal_app_collaborator ADD COLUMN role text;
UPDATE _portal_app_collaborator SET role = 'owner';
ALTER TABLE _portal_app_collaborator ALTER COLUMN role SET NOT NULL;

-- +migrate Down

ALTER TABLE _portal_app_collaborator DROP COLUMN role;
