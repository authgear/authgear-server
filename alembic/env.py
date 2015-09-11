from __future__ import with_statement

from logging.config import fileConfig
import io

from alembic import context
import psycopg2.extensions
from sqlalchemy import create_engine, pool

# this is the Alembic Config object, which provides
# access to the values within the .ini file in use.
config = context.config

# Set up logger, content presented here is copied from
# the auto-generated alembic.ini
fileConfig(io.StringIO("""
# Logging configuration
[loggers]
keys = root,sqlalchemy,alembic

[handlers]
keys = console

[formatters]
keys = generic

[logger_root]
level = WARN
handlers = console
qualname =

[logger_sqlalchemy]
level = WARN
handlers =
qualname = sqlalchemy.engine

[logger_alembic]
level = INFO
handlers =
qualname = alembic

[handler_console]
class = StreamHandler
args = (sys.stderr,)
level = NOTSET
formatter = generic

[formatter_generic]
format = %(levelname)-5.5s [%(name)s] %(message)s
datefmt = %H:%M:%S
"""))

# add your model's MetaData object here
# for 'autogenerate' support
# from myapp import mymodel
# target_metadata = mymodel.Base.metadata
target_metadata = None

# other values from the config, defined by the needs of env.py,
# can be acquired:
# my_important_option = config.get_main_option("my_important_option")
# ... etc.


def run_migrations_offline():
    """Run migrations in 'offline' mode.

    This configures the context with just a URL
    and not an Engine, though an Engine is acceptable
    here as well.  By skipping the Engine creation
    we don't even need a DBAPI to be available.

    Calls to context.execute() here emit the given string to the
    script output.

    """
    url = config.get_main_option("sqlalchemy.url")
    context.configure(
        url=url, target_metadata=target_metadata, literal_binds=True)

    with context.begin_transaction():
        context.run_migrations()


def run_migrations_online():
    """Run migrations in 'online' mode.

    In this scenario we need to create an Engine
    and associate a connection with the context.

    """
    url = config.get_section('db')['option']
    connectable = create_engine(
        url,
        poolclass=pool.NullPool)

    with connectable.connect() as connection:
        appname = config.get_section('app').get('name')
        if not appname:
            raise ValueError('Empty app.name')

        schema = app_schema(appname)
        schema_existed = connection.execute(
            'SELECT EXISTS(SELECT 1 FROM information_schema.schemata '
            'WHERE schema_name = %s)', schema).scalar()
        if not schema_existed:
            print('Creating schema', schema)
            connection.execute('CREATE SCHEMA %s', Identifier(schema))

        connection.execute('SET search_path TO %s, public', Identifier(schema))

        context.configure(
            version_table='_version',
            connection=connection,
            target_metadata=target_metadata,
        )

        prepare_context(context)

        with context.begin_transaction():
            context.run_migrations()


# prepare the alembic context with reusable functions in migration
# not sure it's the right way to do it though
def prepare_context(context):
    context.Identifier = Identifier


def app_schema(appname):
    return 'app_' + appname.lower().replace('.', '_')


class Identifier(str):
    pass


class IdentifierAdapter:

    def __init__(self, s):
        self.s = s

    def prepare(self, conn):
        self.conn = conn

    def getquoted(self):
        return '"%s"' % self.s.replace('"', '""')

psycopg2.extensions.register_adapter(Identifier, IdentifierAdapter)

if context.is_offline_mode():
    run_migrations_offline()
else:
    run_migrations_online()
