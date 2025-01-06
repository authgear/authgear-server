# When you change the version, you also need to change postgresql-server-dev-VERSION below.
FROM postgres:16-bookworm

# Google Cloud SQL PostgreSQL 16 supports up to 5.0.1.
# https://cloud.google.com/sql/docs/postgres/extensions#:~:text=PostgreSQL%20versions%2014%20and%20later%20support%20version%205.0.1%20while%20PostgreSQL%20versions%20that%20are%20earlier%20than%2014%20support%20only%20up%20to%20version%204.7.4.
# So we use that version here.
ENV PARTMAN_VERSION=5.0.1

# Google Cloud SQL PostgreSQL 16 SHOULD BE using pgAudit 16.0
# See https://github.com/pgaudit/pgaudit/releases
# But Google does not say that explicitly.
# https://cloud.google.com/sql/docs/postgres/extensions
#
# pgaudit additionally requires itself to be pre-loaded at server start.
# This is done with a custom postgresql.conf we mount in docker-compose.yaml.
# See shared_preload_libraries in postgresql.conf for details.
#
# Here are the results I got when testing with pgaudit locally.
#
# pgaudit supports 5 use cases.
#
# 1. Global (postgresql.conf or ALTER SYSTEM ... SET)
# I suspect this correspond to [Configure auditing for all database operations on the instance](https://cloud.google.com/sql/docs/postgres/pg-audit#configuring)
#
# 2. Database level (ALTER DATABASE ... SET)
# This corresponds to [Configure auditing for a specific database](https://cloud.google.com/sql/docs/postgres/pg-audit#configuring-for-database)
# When you run `ALTER DATABASE somedb SET pgaudit.log = 'something';`, and then run `SELECT * FROM pg_settings WHERE name = 'pgaudit.log';`
# you will see no change at all.
# You need to restart the database to see the effect.
#
# 3. Role level (ALTER ROLE ... SET)
# This corresponds to [Configure auditing for a database user](https://cloud.google.com/sql/docs/postgres/pg-audit#configuring-for-given-user)
# In my local testing, it takes effect for NEW CONNECTIONS only, but no restart is required.
#
# 4. Session level (SET pgaudit.log = ...)
# This is not very useful because it is up to the program to opt-in audit log. So it is not mentioned in Google Cloud SQL documentation.
# In my local testing, it takes effect immediately.
#
# 5. Object level
# This corresponds to [Configure auditing for a relation](https://cloud.google.com/sql/docs/postgres/pg-audit#configuring-for-relation)
# This is troublesome because it requires manually configuration for each new table, which is likely to be missed.
#
# pgaudit defines the following classes of statements, which can be used in pgaudit.log.
# - READ: SELECT or COPY.
# - WRITE: INSERT, UPDATE, DELETE, etc.
# - FUNCTION: DO.
# - ROLE: GRANT, REVOKE, CREATE/ALTER/DROP ROLE.
# - DDL: All DDL not in ROLE.
# - MISC: BEGIN, ROLLBACK, COMMIT, SET, VACUUM, etc.
# - MISC_SET: SET ROLE, etc.
# - ALL: all of above.
#
# In my local testing, the statement is logged immediately when it is executed, regardless of transaction.
# So merely tracing the audit logs cannot reveal actual changes to the database.
# The audit logs only show literally what statements were executed.
ENV PGAUDIT_VERSION=16.0

RUN apt-get update && apt-get install -y \
	unzip \
	build-essential \
	wget \
	# pgaudit requires <postgres.h> to build
	postgresql-server-dev-16 \
	# And <postgres.h> includes <gssapi/gssapi.h>
	libkrb5-dev \
	&& rm -rf /var/lib/apt/lists/*

RUN \
  wget https://github.com/pgpartman/pg_partman/archive/v${PARTMAN_VERSION}.zip \
    -O pg_partman-${PARTMAN_VERSION}.zip && \
  unzip pg_partman-${PARTMAN_VERSION}.zip && \
  cd pg_partman-${PARTMAN_VERSION} && \
  make NO_BGW=1 install

RUN \
  wget https://github.com/pgaudit/pgaudit/archive/${PGAUDIT_VERSION}.zip \
    -O pgaudit-${PGAUDIT_VERSION}.zip && \
  unzip pgaudit-${PGAUDIT_VERSION}.zip && \
  cd pgaudit-${PGAUDIT_VERSION} && \
  make install USE_PGXS=1
