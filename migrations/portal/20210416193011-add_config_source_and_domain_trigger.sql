-- +migrate Up

-- +migrate StatementBegin
CREATE FUNCTION notify_config_source_change() RETURNS TRIGGER AS $$
DECLARE
  record RECORD;
BEGIN
  IF (TG_OP = 'DELETE') THEN
    record := OLD;
  ELSE
    record := NEW;
  END IF;
  PERFORM pg_notify('config_source_change', record.app_id);
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

CREATE TRIGGER notify_config_source_change
AFTER INSERT OR UPDATE OR DELETE
ON _portal_config_source
FOR ROW
EXECUTE FUNCTION notify_config_source_change();

-- +migrate StatementBegin
CREATE FUNCTION notify_domain_change() RETURNS TRIGGER AS $$
  DECLARE
    record RECORD;
  BEGIN
    IF (TG_OP = 'DELETE') THEN
      record := OLD;
    ELSE
      record := NEW;
    END IF;
    PERFORM pg_notify('domain_change', record.domain);
    RETURN NULL;
  END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

CREATE TRIGGER notify_domain_change
AFTER INSERT OR UPDATE OR DELETE
ON _portal_domain
FOR ROW
EXECUTE FUNCTION notify_domain_change();

-- +migrate Down

DROP TRIGGER notify_config_source_change on _portal_config_source;
DROP FUNCTION notify_config_source_change;

DROP TRIGGER notify_domain_change on _portal_domain;
DROP FUNCTION notify_domain_change;
