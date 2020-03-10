ALTER TABLE plan ADD is_default boolean NOT NULL DEFAULT FALSE;

INSERT INTO
    "plan"("id","created_at","updated_at","name","auth_enabled", "is_default")
VALUES
    ('851fe550-e9d5-4d40-9ee1-2fd1e5f70ec7',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,'default',TRUE,TRUE);
