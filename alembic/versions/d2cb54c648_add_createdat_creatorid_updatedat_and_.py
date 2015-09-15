"""Add _created_at, _created_by, _updated_at and _updated_by to record tables

Revision ID: d2cb54c648
Revises: 48b961caa
Create Date: 2015-09-15 14:57:56.135610

"""

# revision identifiers, used by Alembic.
revision = 'd2cb54c648'
down_revision = '48b961caa'
branch_labels = None
depends_on = None

from alembic import op, context
import sqlalchemy as sa


def upgrade():
    for tablename in context.record_tablenames():
        op.add_column(tablename, sa.Column('_created_at', sa.DateTime(timezone=False)))
        op.add_column(tablename, sa.Column('_updated_at', sa.DateTime(timezone=False)))

        op.execute(
            "UPDATE %s SET _created_at = now() at time zone 'utc', _updated_at = now() at time zone 'utc'" %
            context.quotedIdentifier(tablename))
        op.alter_column(tablename, '_created_at', nullable=False)
        op.alter_column(tablename, '_updated_at', nullable=False)

        op.add_column(tablename, sa.Column('_created_by', sa.Text))
        op.add_column(tablename, sa.Column('_updated_by', sa.Text))


def downgrade():
    for tablename in context.record_tablenames():
        op.drop_column(tablename, '_created_at')
        op.drop_column(tablename, '_created_by')
        op.drop_column(tablename, '_updated_at')
        op.drop_column(tablename, '_updated_by')
