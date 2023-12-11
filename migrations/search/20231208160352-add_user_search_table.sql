-- +migrate Up
CREATE TABLE _search_user
(
    id                            text PRIMARY KEY,
    app_id                        text                        NOT NULL,
    created_at                    timestamp without time zone NOT NULL,
    updated_at                    timestamp without time zone NOT NULL,
    last_login_at                 timestamp without time zone,
    is_disabled                   boolean                     NOT NULL,
    emails                        jsonb                       NOT NULL DEFAULT '[]'::jsonb,
    email_local_parts             jsonb                       NOT NULL DEFAULT '[]'::jsonb,
    email_domains                 jsonb                       NOT NULL DEFAULT '[]'::jsonb,
    preferred_usernames           jsonb                       NOT NULL DEFAULT '[]'::jsonb,
    phone_numbers                 jsonb                       NOT NULL DEFAULT '[]'::jsonb,
    phone_number_country_codes    jsonb                       NOT NULL DEFAULT '[]'::jsonb,
    phone_number_national_numbers jsonb                       NOT NULL DEFAULT '[]'::jsonb,
    oauth_subject_ids             jsonb                       NOT NULL DEFAULT '[]'::jsonb,
    gender                        text                        NOT NULL,
    zoneinfo                      text                        NOT NULL,
    locale                        text                        NOT NULL,
    postal_code                   text                        NOT NULL,
    country                       text                        NOT NULL,
    details                       jsonb                       NOT NULL DEFAULT '{}'::jsonb,
    details_tsvector              tsvector
      GENERATED ALWAYS AS  (jsonb_to_tsvector('english', details, '["string", "numeric"]')) STORED
);
CREATE INDEX _search_user_app_id ON _search_user (app_id);
CREATE INDEX _search_user_app_id_created_at ON _search_user (app_id, created_at);
CREATE INDEX _search_user_app_id_last_login_at ON _search_user (app_id, last_login_at);
CREATE INDEX _search_user_details_tsvector ON _search_user USING GIN (details_tsvector);

-- +migrate Down
DROP TABLE _search_user;
