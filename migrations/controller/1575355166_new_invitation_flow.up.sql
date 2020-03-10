ALTER TABLE invitation DROP COLUMN "code";

ALTER TABLE invitation ADD CONSTRAINT "invitation_email_app_id_uniq" UNIQUE ("email", "app_id");
