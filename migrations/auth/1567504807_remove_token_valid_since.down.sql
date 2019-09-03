-- Put downgrade SQL here
ALTER TABLE _core_user ADD COLUMN "token_valid_since" timestamp without time zone;
