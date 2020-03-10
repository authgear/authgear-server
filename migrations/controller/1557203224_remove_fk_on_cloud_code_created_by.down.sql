-- Put downgrade SQL here
ALTER TABLE cloud_code DROP COLUMN created_by;
ALTER TABLE cloud_code ADD COLUMN created_by UUID REFERENCES app (id);
