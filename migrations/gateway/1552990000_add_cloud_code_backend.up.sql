-- Put upgrade SQL here
ALTER TABLE cloud_code ADD backend_url TEXT NOT NULL DEFAULT '';
