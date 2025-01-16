-- +migrate Up

ALTER TABLE _auth_user ADD COLUMN "mfa_grace_period_end_at" TIMESTAMP WITHOUT TIME ZONE;

-- +migrate Down

ALTER TABLE _auth_user DROP COLUMN "mfa_grace_period_end_at";

