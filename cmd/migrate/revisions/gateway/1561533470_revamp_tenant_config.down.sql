-- Put downgrade SQL here
ALTER TABLE config ALTER COLUMN "config_old" DROP DEFAULT;
ALTER TABLE config DROP COLUMN "config";
ALTER TABLE config RENAME COLUMN "config_old" TO "config";
