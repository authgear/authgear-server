"""add_user_role

Revision ID: 41af1c8d394
Revises: 30d0a626888
Create Date: 2016-01-12 18:38:10.307911

"""

# revision identifiers, used by Alembic.
revision = '41af1c8d394'
down_revision = '30d0a626888'
branch_labels = None
depends_on = None

from alembic import op
import sqlalchemy as sa


def upgrade():
    """
    SQL That equal to the following
    CREATE TABLE app_name._role (
        id text PRIMARY KEY
    );
    CREATE TABLE app_name._user_role (
        user_id text REFERENCES app_name._user (id) NOT NULL,
        role_id text REFERENCES app_name._role (id) NOT NULL,
        PRIMARY KEY (user_id, role_id)
    );
    UPDATE app_name._version set version_num = '41af1c8d394';
    """
    op.create_table(
        '_role',
        sa.Column('id', sa.Text, primary_key=True),
    )
    op.create_table(
        '_user_role',
        sa.Column('user_id', sa.Text, sa.ForeignKey('_user.id'), nullable=False),
        sa.Column('role_id', sa.Text, sa.ForeignKey('_role.id'), nullable=False),
        sa.UniqueConstraint('user_id', 'role_id'),
    )


def downgrade():
    op.drop_table('_user_role')
    op.drop_table('_role')
