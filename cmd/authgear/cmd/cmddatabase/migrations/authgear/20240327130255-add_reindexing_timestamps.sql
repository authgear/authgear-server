-- +migrate Up
ALTER TABLE _auth_user ADD COLUMN last_indexed_at timestamp without time zone;
ALTER TABLE _auth_user ADD COLUMN require_reindex_after timestamp without time zone;
UPDATE _auth_user SET require_reindex_after = updated_at;

-- +migrate Down
ALTER TABLE _auth_user DROP COLUMN last_indexed_at;
ALTER TABLE _auth_user DROP COLUMN require_reindex_after;
