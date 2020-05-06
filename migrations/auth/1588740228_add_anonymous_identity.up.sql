CREATE TABLE _auth_identity_anonymous(
    identity_id TEXT NOT NULL REFERENCES _auth_identity(id) PRIMARY KEY,
    app_id TEXT NOT NULL,
    key_id TEXT NOT NULL,
    key JSONB NOT NULL,
    UNIQUE (app_id, key_id)
);
