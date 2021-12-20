-- +migrate Up

ALTER TABLE _auth_identity_anonymous
  ALTER COLUMN key_id DROP NOT NULL,
  ALTER COLUMN key DROP NOT NULL;

-- +migrate Down
-- migration down may fail if there are records that the keys are empty
ALTER TABLE _auth_identity_anonymous
    ALTER COLUMN key_id SET NOT NULL,
    ALTER COLUMN key SET NOT NULL;
