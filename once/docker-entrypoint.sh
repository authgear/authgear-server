#!/bin/bash
# We use /bin/bash instead of /bin/sh because we need process substitution.
# See https://www.shellcheck.net/wiki/SC3001

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

check_MINIO_ROOT_PASSWORD_is_set() {
	if [ -z "$MINIO_ROOT_PASSWORD" ]; then
		printf 1>&2 "MINIO_ROOT_PASSWORD must be set.\n"
		exit 1
	fi
}

check_AUTHGEAR_ONCE_environment_variables_are_set() {
	if [ -z "$AUTHGEAR_ONCE_ADMIN_USER_EMAIL" ]; then
		printf 1>&2 "AUTHGEAR_ONCE_ADMIN_USER_EMAIL must be set.\n"
		exit 1
	fi
	if [ -z "$AUTHGEAR_ONCE_ADMIN_USER_PASSWORD" ]; then
		printf 1>&2 "AUTHGEAR_ONCE_ADMIN_USER_PASSWORD must be set.\n"
		exit 1
	fi
	if [ "$(printf "%s" "$AUTHGEAR_ONCE_ADMIN_USER_PASSWORD" | wc -c)" -lt 8 ]; then
		printf 1>&2 "AUTHGEAR_ONCE_ADMIN_USER_PASSWORD must be at least 8 characters long.\n"
		exit 1
	fi
}

check_http_origin_is_set() {
	name="AUTHGEAR_HTTP_ORIGIN_$1"
	value="${!name}"

	if [ -z "$value" ]; then
		printf 1>&2 "%s must be set. This will be used to configure nginx and certbot.\n" "$name"
		exit 1
	fi
}

check_http_origin_not_equal_to_each_other() {
	name1="AUTHGEAR_HTTP_ORIGIN_$1"
	name2="AUTHGEAR_HTTP_ORIGIN_$2"
	value1="${!name1}"
	value2="${!name2}"

	if [ "$value1" = "$value2" ]; then
		printf 1>&2 "%s must not equal %s.\n" "$name1" "$name2"
		exit 1
	fi
}

docker_postgresql_create_database_directories() {
	sudo mkdir -p "$PGDATA"
	sudo chmod 0700 "$PGDATA"
	sudo mkdir -p /var/run/postgresql
	sudo chmod 0775 /var/run/postgresql

	sudo find "$PGDATA" \! -user "authgear" -exec chown "authgear":"authgear" '{}' +
	sudo find /var/run/postgresql \! -user "authgear" -exec chown "authgear":"authgear" '{}' +
}

docker_postgresql_initdb() {
	# This is the most secure method supported by PostgreSQL 16.
	# See https://www.postgresql.org/docs/16/auth-password.html
	auth_method="scram-sha-256"

	# Google Cloud SQL PostgreSQL uses libc as the locale provider.
	# So we just follow it.
	# Since --locale-provider=libc is the default, we need not specify it.
	# NOTE: --pwfile uses process substitution.
	initdb \
		--username="authgear" \
		--pwfile=<(printf "%s\n" "$POSTGRES_PASSWORD") \
		--encoding="UTF8" \
		--auth-host="$auth_method" \
		--auth-local="$auth_method"
}

docker_postgresql_temp_server_start() {
	pg_ctl start --wait
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

	# Connect to the database `postgres` that must exists.
	db_exists="$(docker_postgresql_psql \
		--dbname postgres \
		--set db="authgear" \
		--tuples-only <<-'EOSQL'
SELECT 1 FROM pg_database WHERE datname = :'db';
	EOSQL
	)"
	if [ -z "$db_exists" ]; then
		docker_postgresql_psql \
			--dbname postgres \
			--set db="authgear" <<-'EOSQL'
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

	sudo find /var/lib/redis/data \! -user "authgear" -exec chown "authgear:authgear" '{}' +
	sudo find /var/run/redis \! -user "authgear" -exec chown "authgear:authgear" '{}' +
}

docker_redis_write_acl_file() {
	printf "user default on +@all ~* >%s\n" "$REDIS_PASSWORD" > /var/run/redis/users.acl
}

docker_nginx_create_directories() {
	# This is the default location of --prefix
	sudo mkdir -p /usr/share/nginx
	sudo chmod 0755 /usr/share/nginx
	# Ubuntu nginx-common installs files to here.
	sudo mkdir -p /etc/nginx
	sudo chmod 0755 /etc/nginx
	# When I run `nginx -t`, it tries to write files to /var/lib/nginx.
	sudo mkdir -p /var/lib/nginx
	sudo chmod 0755 /var/lib/nginx
	# The directory to contain the pid file.
	sudo mkdir -p /var/run/nginx
	sudo chmod 0755 /var/run/nginx
	# The directory to contain the log file.
	sudo mkdir -p /var/log/nginx
	sudo chmod 0755 /var/log/nginx
	# The directory to write temp files.
	sudo mkdir -p /tmp/nginx
	sudo chmod 0755 /tmp/nginx

	# There is a broken symlink /usr/share/nginx/modules
	# I do not know why.
	sudo rm -f /usr/share/nginx/modules

	sudo find /usr/share/nginx \! -user "authgear" -exec chown "authgear:authgear" '{}' +
	sudo find /etc/nginx \! -user "authgear" -exec chown "authgear:authgear" '{}' +
	sudo find /var/lib/nginx \! -user "authgear" -exec chown "authgear:authgear" '{}' +
	sudo find /var/run/nginx \! -user "authgear" -exec chown "authgear:authgear" '{}' +
	sudo find /var/log/nginx \! -user "authgear" -exec chown "authgear:authgear" '{}' +
	sudo find /tmp/nginx \! -user "authgear" -exec chown "authgear:authgear" '{}' +
}

docker_nginx_render_server_block() {
	name="AUTHGEAR_HTTP_ORIGIN_$1"
	value="${!name}"

	if [ -z "$value" ]; then
		printf 1>&2 "%s must be set. This will be used to configure nginx and certbot.\n" "$name"
		exit 1
	fi

	scheme=""
	without_scheme=""
	case "$value" in
	"http://"*)
		scheme="http"
		without_scheme="${value#http://}"
		;;
	"https://"*)
		scheme="https"
		without_scheme="${value#https://}"
		;;
	*)
		;;
	esac
	if [ -z "$without_scheme" ]; then
		printf 1>&2 "%s must start with http:// or https://.\n" "$name"
		exit 1
	fi

	contains_port="false"
	contains_path="false"
	contains_query="false"
	contains_fragment="false"
	case "$without_scheme" in
	*:*)
		contains_port="true"
		;;
	*/*)
		contains_path="true"
		;;
	*'?'*)
		contains_query="true"
		;;
	*'#'*)
		contains_fragment="true"
		;;
	*)
		;;
	esac
	if [ "$contains_port" = "true" ]; then
		printf 1>&2 "%s must not contain port.\n" "$name"
		exit 1
	fi
	if [ "$contains_path" = "true" ]; then
		printf 1>&2 "%s must not contain path.\n" "$name"
		exit 1
	fi
	if [ "$contains_query" = "true" ]; then
		printf 1>&2 "%s must not contain query.\n" "$name"
		exit 1
	fi
	if [ "$contains_fragment" = "true" ]; then
		printf 1>&2 "%s must not contain fragment.\n" "$name"
		exit 1
	fi

	sed -i -E "s,__AUTHGEAR_HTTP_HOST_${1}__,${without_scheme}," /etc/nginx/nginx.conf
	sed -i -E "s,__AUTHGEAR_HTTP_EXPECTED_SCHEME_${1}__,${scheme}," /etc/nginx/nginx.conf
}

docker_nginx_check_environment_variables() {
	check_http_origin_is_set "PORTAL"
	check_http_origin_is_set "ACCOUNTS"
	check_http_origin_is_set "PROJECT"

	check_http_origin_not_equal_to_each_other "PORTAL" "ACCOUNTS"
	check_http_origin_not_equal_to_each_other "PORTAL" "PROJECT"
	check_http_origin_not_equal_to_each_other "ACCOUNTS" "PROJECT"
}

docker_nginx_create_nginx_conf() {
	cp /etc/nginx/nginx.conf.sample /etc/nginx/nginx.conf
	docker_nginx_render_server_block "PORTAL"
	docker_nginx_render_server_block "ACCOUNTS"
	docker_nginx_render_server_block "PROJECT"
}

docker_nginx_create_fake_certificate() {
	openssl req \
		-batch \
		-x509 \
		-newkey rsa:4096 \
		-noenc \
		-keyout /etc/nginx/fake_certficate.key \
		-out /etc/nginx/fake_certficate.crt \
		-subj "/CN=invalid.invalid/OU=Fake certificate generated by docker-entrypoint.sh" \
		-days 365
}

docker_certbot_create_directories() {
	sudo mkdir -p /var/lib/letsencrypt
	sudo mkdir -p /var/log/letsencrypt
	sudo mkdir -p /etc/letsencrypt

	sudo find /var/lib/letsencrypt \! -user "authgear" -exec chown "authgear:authgear" '{}' +
	sudo find /var/log/letsencrypt \! -user "authgear" -exec chown "authgear:authgear" '{}' +
	sudo find /etc/letsencrypt \! -user "authgear" -exec chown "authgear:authgear" '{}' +
}

docker_certbot_create_cli_ini() {
	# certbot stores its data (like certificates, accounts) in --config-dir
	# It also reads its config (cli.ini) in --config-dir.
	# If we naively volume mount /etc/letsencrypt, /etc/letsencrypt/cli.ini will be obscured by the mount.
	# Therefore, when we build the image, we copy the original /etc/letsencrypt/cli.ini to /home/authgear/certbot.ini,
	# and we always write a fresh /etc/letsencrypt/cli.ini.
	cli_ini="/etc/letsencrypt/cli.ini"
	cp /home/authgear/certbot.ini.example "$cli_ini"
	sed -E -i 's,^#?\s*(max-log-backups)\s+.*,\1 = 10,' "$cli_ini"
	sed -E -i 's,^#?\s*(preconfigured-renewal)\s+.*,\1 = False,' "$cli_ini"

	if ! grep -E '^max-log-backups = 10' "$cli_ini" 1>/dev/null; then
		printf 1>&2 "failed to set max-log-backups in %s\n" "$cli_ini"
		exit 1
	fi

	if ! grep -E '^preconfigured-renewal = False' "$cli_ini" 1>/dev/null; then
		printf 1>&2 "failed to set preconfigured-renewal in %s\n" "$cli_ini"
		exit 1
	fi
}

docker_minio_create_directories() {
	sudo mkdir -p /var/lib/minio/data
	sudo chmod 0700 /var/lib/minio/data

	sudo find /var/lib/minio/data \! -user "authgear" -exec chown "authgear:authgear" '{}' +
}

docker_minio_create_buckets() {
	# We need to start the server temporarily.
	minio server /var/lib/minio/data &
	minio_pid=$!

	# 1 second should be enough. It starts very fast.
	sleep 1
	mcli alias set local http://localhost:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD"
	mcli mb --ignore-existing local/images
	mcli mb --ignore-existing local/userexport

	kill -SIGTERM "$minio_pid"
}

docker_tls_update_ca_certificates() {
	# Make the certificates in /usr/local/share/ca-certificates take effect.
	sudo update-ca-certificates
}

docker_dns_update_etc_hosts() {
	echo '127.0.0.1 accounts.projects.authgear' | sudo tee -a /etc/hosts >/dev/null
	echo '127.0.0.1 project.projects.authgear' | sudo tee -a /etc/hosts >/dev/null
}

docker_authgear_source_env() {
	# It is put at ~/.bashrc so that
	# `docker compose exec THE_SERVICE bash` will also source this file.
	source "/home/authgear/.bashrc"
}

docker_authgear_run_database_migrations() {
	docker_postgresql_temp_server_start

	authgear database migrate up
	authgear audit database migrate up
	authgear images database migrate up
	authgear search database migrate up
	authgear-portal database migrate up

	docker_postgresql_temp_server_stop
}

docker_authgear_create_deployment_runtime_directory() {
	mkdir -p /home/authgear/authgear_deployment_runtime

	cat > /home/authgear/authgear_deployment_runtime/authgear.secrets.yaml <<EOF
secrets:
- key: db
  data:
    database_schema: "$DATABASE_SCHEMA"
    database_url: "$DATABASE_URL"
- key: audit.db
  data:
    database_schema: "$AUDIT_DATABASE_SCHEMA"
    database_url: "$AUDIT_DATABASE_URL"
- key: search.db
  data:
    database_schema: "$SEARCH_DATABASE_SCHEMA"
    database_url: "$SEARCH_DATABASE_URL"
- key: redis
  data:
    redis_url: "$REDIS_URL"
- key: analytic.redis
  data:
    redis_url: "$ANALYTIC_REDIS_URL"
EOF
}

docker_authgear_init() {
	docker_wrapper &
	wrapper_pid=$!
	# Wait 2 seconds for the server to start.
	sleep 2


	app_id_accounts="accounts"
	init_output_accounts="$(mktemp -d)"
	authgear init --interactive=false \
		--purpose=portal \
		--for-helm-chart=true \
		--app-id="$app_id_accounts" \
		--public-origin="$AUTHGEAR_HTTP_ORIGIN_ACCOUNTS" \
		--portal-origin="$AUTHGEAR_HTTP_ORIGIN_PORTAL" \
		--portal-client-id=portal \
		--phone-otp-mode=sms \
		--disable-email-verification=true \
		--search-implementation=postgresql \
		-o "$init_output_accounts"
	authgear-portal internal configsource create "$init_output_accounts"
	host_accounts="$(echo "$AUTHGEAR_HTTP_ORIGIN_ACCOUNTS" | awk -F '://' '{ print $2 }')"
	authgear-portal internal domain create-custom "$app_id_accounts" --domain "$host_accounts" --apex-domain "$host_accounts"
	rm -r "$init_output_accounts"

	app_id_project="project"
	init_output_project="$(mktemp -d)"
	authgear init --interactive=false \
		--purpose=project \
		--for-helm-chart=true \
		--app-id="$app_id_project" \
		--public-origin="$AUTHGEAR_HTTP_ORIGIN_PROJECT" \
		--phone-otp-mode=sms \
		--disable-email-verification=true \
		--search-implementation=postgresql \
		-o "$init_output_project"
	authgear-portal internal configsource create "$init_output_project"
	host_project="$(echo "$AUTHGEAR_HTTP_ORIGIN_PROJECT" | awk -F '://' '{ print $2 }')"
	authgear-portal internal domain create-custom "$app_id_project" --domain "$host_project" --apex-domain "$host_project"
	rm -r "$init_output_project"

	# Create default domain after all projects have been created.
	authgear-portal internal domain create-default --default-domain-suffix '.projects.authgear'

	query_file="$(mktemp)"
	cat >"$query_file" <<'EOF'
mutation createUser($email: String!, $password: String!) {
  createUser(input: {
    definition: {
      loginID: {
        key: "email"
        value: $email
      }
    }
    password: $password
  }) {
    user {
      id
    }
  }
}
EOF

	query_output="$(mktemp)"
	authgear internal admin-api invoke \
		--app-id accounts \
		--endpoint "http://localhost:3002" \
		--host "accounts.projects.authgear" \
		--query-file "$query_file" \
		--operation-name "createUser" \
		--variables-json "$(jq -cn --arg email "$AUTHGEAR_ONCE_ADMIN_USER_EMAIL" --arg password "$AUTHGEAR_ONCE_ADMIN_USER_PASSWORD" '{email: $email, password: $password}')" | tee "$query_output"
	decoded_node_id="$(jq <"$query_output" --raw-output '.data.createUser.user.id' | basenc --base64url --decode)"
	raw_id="${decoded_node_id#User:}"

	# Add collaborator to both projects.
	authgear-portal internal collaborator add \
		--app-id "$app_id_accounts" \
		--user-id "$raw_id" \
		--role owner
	authgear-portal internal collaborator add \
		--app-id "$app_id_project" \
		--user-id "$raw_id" \
		--role owner

	kill -SIGTERM "$wrapper_pid"
	# Wait 2 seconds to let the process exits.
	sleep 2
}

main() {
	run_initialization=''

	check_user_is_correct
	check_PGDATA_is_set
	check_LANG_is_set
	check_POSTGRES_PASSWORD_is_set
	check_REDIS_PASSWORD_is_set
	check_MINIO_ROOT_PASSWORD_is_set
	check_AUTHGEAR_ONCE_environment_variables_are_set
	docker_nginx_check_environment_variables

	docker_nginx_create_directories
	docker_nginx_create_fake_certificate
	docker_nginx_create_nginx_conf

	docker_certbot_create_directories
	docker_certbot_create_cli_ini

	docker_postgresql_create_database_directories

	# If this file exists and its size is greater than zero,
	# then we consider the database has initialized.
	if [ -s "$PGDATA/PG_VERSION" ]; then
		printf 1>&2 "PostgreSQL database directory (%s) seems initialized. Skipping initialization.\n" "$PGDATA"
	else
		run_initialization=1
		docker_postgresql_initdb

		# initdb will create the given database user (--username), and
		# the database `postgres`, `template1` and `template0`.
		# It does not provide an option to create more additional databases.
		# Thus we need to do that in a separate step.
		docker_postgresql_create_database
	fi

	docker_redis_create_directories
	docker_redis_write_acl_file

	docker_minio_create_directories
	docker_minio_create_buckets

	docker_tls_update_ca_certificates

	docker_dns_update_etc_hosts

	docker_authgear_source_env
	docker_authgear_run_database_migrations
	docker_authgear_create_deployment_runtime_directory
	if [ -n "$run_initialization" ]; then
		docker_authgear_init
		exit_status="$?"
		if [ "$exit_status" -ne 0 ]; then
			printf 1>&2 "Deleting PostgreSQL database directory (%s) due to error.\n" "$PGDATA"
			# We cannot remove "$PGDATA" itself because it is a volume mount.
			sudo find "$PGDATA" -mindepth 1 -delete
			exit "$exit_status"
		fi
	fi

	# Replace this process with the given arguments.
	exec "$@"
}

main "$@"
