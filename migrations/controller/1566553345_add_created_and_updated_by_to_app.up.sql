ALTER TABLE app ADD COLUMN "created_by" text REFERENCES _core_user(id);
UPDATE app
    SET created_by = app_user.user_id
FROM
    app_user
WHERE
    app_user.app_id = app.id;

ALTER TABLE app ALTER COLUMN "created_by" SET NOT NULL;

ALTER TABLE app ADD COLUMN "updated_by" text REFERENCES _core_user(id);
UPDATE app
    SET updated_by = app_user.user_id
FROM
    app_user
WHERE
    app_user.app_id = app.id;

ALTER TABLE app ALTER COLUMN "updated_by" SET NOT NULL;
