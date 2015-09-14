"""empty message

Revision ID: 48b961caa
Revises: 
Create Date: 2015-09-09 18:32:33.468117

"""

# revision identifiers, used by Alembic.
revision = '48b961caa'
down_revision = None
branch_labels = None
depends_on = None

from alembic import op


def upgrade():
    for statement in STATEMENTS:
        op.execute(statement)


def downgrade():
    curr_schema = op.get_bind().execute('SELECT current_schema').scalar()
    raise NotImplementedError(
        'migration %s cannot be downgraded to, '
        'run DROP SCHEMA %s CASCADE directly' % (revision, curr_schema))

STATEMENTS = [
    """
    CREATE TABLE IF NOT EXISTS public.pending_notification (
        id SERIAL NOT NULL PRIMARY KEY,
        op text NOT NULL,
        appname text NOT NULL,
        recordtype text NOT NULL,
        record jsonb NOT NULL
    );
    """,
    """
    CREATE OR REPLACE FUNCTION public.notify_record_change() RETURNS TRIGGER AS $$
        DECLARE
            affected_record RECORD;
            inserted_id integer;
        BEGIN
            IF (TG_OP = 'DELETE') THEN
                affected_record := OLD;
            ELSE
                affected_record := NEW;
            END IF;
            INSERT INTO pending_notification (op, appname, recordtype, record)
                VALUES (TG_OP, TG_TABLE_SCHEMA, TG_TABLE_NAME, row_to_json(affected_record)::jsonb)
                RETURNING id INTO inserted_id;
            PERFORM pg_notify('record_change', inserted_id::TEXT);
            RETURN affected_record;
        END;
    $$ LANGUAGE plpgsql;
    """,
    """
    CREATE TABLE _user (
        id text PRIMARY KEY,
        email text,
        password text,
        auth jsonb
    );
    """,
    """
    CREATE TABLE _asset (
        id text PRIMARY KEY,
        content_type text NOT NULL,
        size bigint NOT NULL
    );
    """,
    """
    CREATE TABLE _device (
        id text PRIMARY KEY,
        user_id text REFERENCES _user (id),
        type text NOT NULL,
        token text NOT NULL,
        last_registered_at timestamp without time zone NOT NULL,
        UNIQUE (user_id, type, token)
    );
    """,
    """
    CREATE INDEX ON _device (token, last_registered_at);
    """,
    """
    CREATE TABLE _subscription (
        id text NOT NULL,
        user_id text NOT NULL,
        device_id text REFERENCES _device (id) ON DELETE CASCADE NOT NULL,
        type text NOT NULL,
        notification_info jsonb,
        query jsonb,
        PRIMARY KEY(user_id, device_id, id)
    );
    """,
    """
    CREATE TABLE _friend (
        left_id text NOT NULL,
        right_id text REFERENCES _user (id) NOT NULL,
        PRIMARY KEY(left_id, right_id)
    );
    """,
    """
    CREATE TABLE _follow (
        left_id text NOT NULL,
        right_id text REFERENCES _user (id) NOT NULL,
        PRIMARY KEY(left_id, right_id)
    );
    """,
]
