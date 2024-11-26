-- +migrate Up
CREATE TABLE _search_user
(
    id                            text PRIMARY KEY,
    app_id                        text                        NOT NULL,
    app_ids                       text[]
      GENERATED ALWAYS AS (ARRAY[app_id]) STORED,
    created_at                    timestamp without time zone NOT NULL,
    updated_at                    timestamp without time zone NOT NULL,
    last_login_at                 timestamp without time zone,
    is_disabled                   boolean                     NOT NULL,
    emails                        text[]                      NOT NULL DEFAULT '{}',
    email_local_parts             text[]                      NOT NULL DEFAULT '{}',
    email_domains                 text[]                      NOT NULL DEFAULT '{}',
    preferred_usernames           text[]                      NOT NULL DEFAULT '{}',
    phone_numbers                 text[]                      NOT NULL DEFAULT '{}',
    phone_number_country_codes    text[]                      NOT NULL DEFAULT '{}',
    phone_number_national_numbers text[]                      NOT NULL DEFAULT '{}',
    oauth_subject_ids             text[]                      NOT NULL DEFAULT '{}',
    gender                        text[]                      NOT NULL DEFAULT '{}',
    zoneinfo                      text[]                      NOT NULL DEFAULT '{}',
    locale                        text[]                      NOT NULL DEFAULT '{}',
    postal_code                   text[]                      NOT NULL DEFAULT '{}',
    country                       text[]                      NOT NULL DEFAULT '{}',
    role_keys                     text[]                      NOT NULL DEFAULT '{}',
    group_keys                    text[]                      NOT NULL DEFAULT '{}',
    details                       jsonb                       NOT NULL DEFAULT '{}'::jsonb,
    details_tsvector              tsvector
      GENERATED ALWAYS AS (jsonb_to_tsvector('simple', details, '["string", "numeric"]')) STORED
);
CREATE INDEX _search_user_app_id ON _search_user (app_id);
CREATE INDEX _search_user_app_id_created_at ON _search_user (app_id, created_at);
CREATE INDEX _search_user_app_id_last_login_at ON _search_user (app_id, last_login_at);
CREATE INDEX _search_user_gin ON _search_user USING GIN (
  app_ids,
  emails,
  email_domains,
  email_local_parts,
  preferred_usernames,
  phone_numbers,
  phone_number_country_codes,
  phone_number_national_numbers,
  oauth_subject_ids,
  gender,
  zoneinfo,
  locale,
  postal_code,
  country,
  role_keys,
  group_keys,
  details_tsvector);

-- +migrate Down
DROP TABLE _search_user;
