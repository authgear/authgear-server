ALTER TABLE static_deployment ADD COLUMN fallback_page TEXT;
UPDATE static_deployment SET fallback_page = '';
ALTER TABLE static_deployment ALTER COLUMN fallback_page SET NOT NULL;
