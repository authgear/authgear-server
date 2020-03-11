ALTER TABLE invitation ADD COLUMN "code" TEXT;
UPDATE invitation SET "code" = '';
ALTER TABLE invitation ALTER COLUMN "code" SET NOT NULL;

ALTER TABLE invitation DROP CONSTRAINT "invitation_email_app_id_uniq";
