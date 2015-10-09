"""Make _device.token nullable

Revision ID: 51375067b45
Revises: d2cb54c648
Create Date: 2015-10-09 04:07:31.590938

"""

# revision identifiers, used by Alembic.
revision = '51375067b45'
down_revision = 'd2cb54c648'
branch_labels = None
depends_on = None

from alembic import op
import sqlalchemy as sa


def upgrade():
    op.execute("ALTER TABLE _device ALTER COLUMN token DROP NOT NULL; ")


def downgrade():
    op.execute("ALTER TABLE _device ALTER COLUMN token SET NOT NULL; ")
