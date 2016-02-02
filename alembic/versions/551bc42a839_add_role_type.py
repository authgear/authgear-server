"""add_role_type

Revision ID: 551bc42a839
Revises: 41af1c8d394
Create Date: 2016-02-02 17:19:21.998718

"""

# revision identifiers, used by Alembic.
revision = '551bc42a839'
down_revision = '41af1c8d394'
branch_labels = None
depends_on = None

from alembic import op
import sqlalchemy as sa


def upgrade():
    """
    SQL That equal to the following
    ALTER TABLE app_name._role ADD COLUMN by_default boolean DEFAULT FALSE;
    ALTER TABLE app_name._role ADD COLUMN is_admin boolean DEFAULT FALSE;
    UPDATE app_name._version set version_num = '551bc42a839;
    """
    op.add_column('_role', sa.Column('is_admin', sa.Boolean, server_default="FALSE"))
    op.add_column('_role', sa.Column('by_default', sa.Boolean, server_default="FALSE"))


def downgrade():
    op.drop_column('_role', 'is_admin')
    op.drop_column('_role', 'by_default')
