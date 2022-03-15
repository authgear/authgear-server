-- +migrate Up
CREATE TABLE _images_file
(
    id         text  PRIMARY KEY,
    app_id     text  NOT NULL,
    object_id  text  NOT NULL,
    size       int   NOT NULL,
    metadata   jsonb NOT NULL,
    created_at timestamp without time zone NOT NULL,
    UNIQUE (app_id, object_id)
);

-- +migrate Down
DROP TABLE _images_file;
