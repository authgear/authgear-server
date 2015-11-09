"""Add username

Revision ID: 30d0a626888
Revises: 51375067b45
Create Date: 2015-10-29 10:32:03.077400

"""

# revision identifiers, used by Alembic.
revision = '30d0a626888'
down_revision = '51375067b45'
branch_labels = None
depends_on = None

from alembic import op
import sqlalchemy as sa


def upgrade():
    """
    SQL That equal to the following
    ALTER TABLE app_name._user ADD COLUMN username varchar(255);
    ALTER TABLE app_name._user ADD CONSTRAINT '_user_email_key' UNIQUE('email');
    UPDATE app_name._version set version_num = '30d0a626888;
    """
    op.add_column('_user', sa.Column('username', sa.Unicode(255), unique=True))
    op.create_unique_constraint(
        '_user_email_key', '_user', ['email'])


def downgrade():
    op.drop_column('_user', 'username')
    op.drop_constraint(
        '_user_email_key', table_name='_user', type_='unique')
