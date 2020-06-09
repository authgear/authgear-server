-- Put upgrade SQL here
CREATE TABLE _auth_user(
       id text PRIMARY KEY,
       app_id text NOT NULL,
       created_at timestamp without time zone NOT NULL,
       updated_at timestamp without time zone NOT NULL,
       last_login_at timestamp without time zone,
       metadata jsonb
);
INSERT INTO _auth_user
    SELECT
        _core_user.id AS id,
        _core_user.app_id AS app_id,
        _auth_user_profile.created_at AS created_at,
        _auth_user_profile.updated_at AS updated_at,
        _core_user.last_login_at AS last_login_at,
        _auth_user_profile.data AS metadata
    FROM _core_user
        JOIN _auth_user_profile ON (_core_user.id = _auth_user_profile.user_id);

ALTER TABLE _auth_oauth_authorization DROP CONSTRAINT _auth_oauth_authorization_user_id_fkey;
DROP TABLE _core_user;
DROP TABLE _auth_user_profile;
ALTER TABLE _auth_oauth_authorization ADD CONSTRAINT  _auth_oauth_authorization_user_id_fkey FOREIGN KEY(user_id) REFERENCES _auth_user(id);
