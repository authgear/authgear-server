-- +migrate Up
CREATE UNIQUE INDEX _portal_app_collaborator_invitation_code_key ON _portal_app_collaborator_invitation (code);

-- +migrate Down
DROP INDEX _portal_app_collaborator_invitation_code_key;
