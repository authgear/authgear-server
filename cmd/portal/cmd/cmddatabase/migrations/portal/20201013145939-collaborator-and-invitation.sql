-- +migrate Up
CREATE TABLE _portal_app_collaborator
(
	id         text PRIMARY KEY,
	app_id     text                        NOT NULL,
	user_id    text                        NOT NULL,
	created_at timestamp WITHOUT TIME ZONE NOT NULL,
	UNIQUE (app_id, user_id)
);

CREATE TABLE _portal_app_collaborator_invitation
(
	id            text PRIMARY KEY,
	app_id        text                        NOT NULL,
	invited_by    text                        NOT NULL,
	invitee_email text                        NOT NULL,
	code          text                        NOT NULL,
	created_at    timestamp WITHOUT TIME ZONE NOT NULL,
	expire_at     timestamp WITHOUT TIME ZONE NOT NULL,
	UNIQUE (app_id, invitee_email)
);

-- +migrate Down

DROP TABLE _portal_app_collaborator
DROP TABLE _portal_app_collaborator_invitation;
