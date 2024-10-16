-- +migrate Up
CREATE TABLE _auth_authenticator_face_recognition
(
    id                  text PRIMARY KEY REFERENCES _auth_authenticator (id),
    app_id              text                        NOT NULL,
    opencv_fr_person_id text                        NOT NULL
);

-- +migrate Down
DROP TABLE _auth_authenticator_face_recognition;
