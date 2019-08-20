ALTER TABLE _core_user ADD COLUMN app_id TEXT;
UPDATE _core_user SET app_id = '';
ALTER TABLE _core_user ALTER COLUMN app_id SET NOT NULL;
CREATE INDEX _core_user_app_id_idx ON _core_user(app_id);
