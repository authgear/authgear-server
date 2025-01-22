#!/bin/bash
# We use /bin/bash instead of /bin/sh because we need process substitution.
# See https://www.shellcheck.net/wiki/SC3001

set -ex

check_user_is_correct() {
	if [ "$(id -u -n)" != "authgear" ]; then
		printf 1>&2 "docker-entrypoint.sh is supposed to be run with the user authgear.\n"
		exit 1
	fi
}

check_PGDATA_is_set() {
	if [ -z "$PGDATA" ]; then
		printf 1>&2 "PGDATA should be set in the Dockerfile with ENV. Make sure you do not override it.\n"
		exit 1
	fi
}

check_LANG_is_set() {
	if [ "$LANG" != "en_US.utf8" ]; then
		printf 1>&2 "LANG must be set to en_US.utf8\n"
		exit 1
	fi
}

check_POSTGRES_PASSWORD_is_set() {
	if [ -z "$POSTGRES_PASSWORD" ]; then
		printf 1>&2 "POSTGRES_PASSWORD must be set. This will be used to initialize a new database with a password.\n"
		exit 1
	fi
}

check_REDIS_PASSWORD_is_set() {
	if [ -z "$REDIS_PASSWORD" ]; then
		printf 1>&2 "REDIS_PASSWORD must be set. This will be used to set the password for the default user.\n"
		exit 1
	fi
}

docker_postgresql_create_database_directories() {
	sudo mkdir -p "$PGDATA"
	sudo chmod 0700 "$PGDATA"
	sudo mkdir -p /var/run/postgresql
	sudo chmod 0775 /var/run/postgresql

	user="$(id -u -n)"
	sudo find "$PGDATA" \! -user "$user" -exec chown "$user":"$user" '{}' +
	sudo find /var/run/postgresql \! -user "$user" -exec chown "$user":"$user" '{}' +
}

docker_postgresql_initdb() {
	user="$(id -u -n)"

	# This is the most secure method supported by PostgreSQL 16.
	# See https://www.postgresql.org/docs/16/auth-password.html
	auth_method="scram-sha-256"

	# Google Cloud SQL PostgreSQL uses libc as the locale provider.
	# So we just follow it.
	# Since --locale-provider=libc is the default, we need not specify it.
	# NOTE: --pwfile uses process substitution.
	initdb \
		--username="$user" \
		--pwfile=<(printf "%s\n" "$POSTGRES_PASSWORD") \
		--encoding="UTF8" \
		--auth-host="$auth_method" \
		--auth-local="$auth_method"
}

docker_postgresql_temp_server_start() {
	pg_ctl start \
		-o "-c listen_addresses='' -p 5432" \
		--wait
}

docker_postgresql_temp_server_stop() {
	pg_ctl stop \
		--mode fast \
		--wait
}

# docker_postgresql_psql is a wrapper of psql.
docker_postgresql_psql() {
	PGPASSWORD="$POSTGRES_PASSWORD" psql \
		--variable ON_ERROR_STOP=1 \
		--no-psqlrc "$@"
}

docker_postgresql_create_database() {
	docker_postgresql_temp_server_start

	user="$(id -u -n)"
	# Connect to the database `postgres` that must exists.
	db_exists="$(docker_postgresql_psql \
		--dbname postgres \
		--set db="$user" \
		--tuples-only <<-'EOSQL'
SELECT 1 FROM pg_database WHERE datname = :'db';
	EOSQL
	)"
	if [ -z "$db_exists" ]; then
		docker_postgresql_psql \
			--dbname postgres \
			--set db="$user" <<-'EOSQL'
CREATE DATABASE :"db";
	EOSQL
	fi

	docker_postgresql_temp_server_stop
}

docker_redis_create_directories() {
	sudo mkdir -p /var/lib/redis/data
	sudo chmod 0700 /var/lib/redis/data
	sudo mkdir -p /var/run/redis
	sudo chmod 0700 /var/run/redis

	user="$(id -u -n)"
	sudo find /var/lib/redis/data \! -user "$user" -exec chown "$user:$user" '{}' +
	sudo find /var/run/redis \! -user "$user" -exec chown "$user:$user" '{}' +
}

docker_redis_write_acl_file() {
	printf "user default on +@all ~* >%s\n" "$REDIS_PASSWORD" > /var/run/redis/users.acl
}

main() {
	check_user_is_correct
	check_PGDATA_is_set
	check_LANG_is_set
	check_POSTGRES_PASSWORD_is_set
	check_REDIS_PASSWORD_is_set

	docker_postgresql_create_database_directories

	# If this file exists and its size is greater than zero,
	# then we consider the database has initialized.
	if [ -s "$PGDATA/PG_VERSION" ]; then
		printf 1>&2 "PostgreSQL database directory (%s) seems initialized. Skipping initialization.\n" "$PGDATA"
	else
		docker_postgresql_initdb

		# initdb will create the given database user (--username), and
		# the database `postgres`, `template1` and `template0`.
		# It does not provide an option to create more additional databases.
		# Thus we need to do that in a separate step.
		docker_postgresql_create_database
	fi

	docker_redis_create_directories
	docker_redis_write_acl_file

	# Replace this process with the given arguments.
	exec "$@"
}

main "$@"
