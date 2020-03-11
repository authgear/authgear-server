ALTER TABLE static_deployment ADD COLUMN index_file TEXT;
UPDATE static_deployment SET index_file = '';
ALTER TABLE static_deployment ALTER COLUMN index_file SET NOT NULL;
