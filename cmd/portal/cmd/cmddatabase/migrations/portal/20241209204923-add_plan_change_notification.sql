-- +migrate Up

-- +migrate StatementBegin
CREATE FUNCTION notify_plan_change() RETURNS TRIGGER AS $$
DECLARE
  record RECORD;
BEGIN
  IF (TG_OP = 'DELETE') THEN
    record := OLD;
  ELSE
    record := NEW;
  END IF;
  PERFORM pg_notify('plan_change', record.name);
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

CREATE TRIGGER notify_plan_change
AFTER INSERT OR UPDATE OR DELETE
ON _portal_plan
FOR ROW
EXECUTE FUNCTION notify_plan_change();

-- +migrate Down

DROP TRIGGER notify_plan_change on _portal_plan;
DROP FUNCTION notify_plan_change;
