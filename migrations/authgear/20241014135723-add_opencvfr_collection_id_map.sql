-- +migrate Up
CREATE TABLE _auth_authenticator_face_recognition_opencvfr_collection_map
(
    id                      text                     PRIMARY KEY,
    app_id                  text                        NOT NULL,
    opencv_fr_collection_id text                        NOT NULL
);

-- +migrate Down
DROP TABLE _auth_authenticator_face_recognition_opencvfr_collection_map;
