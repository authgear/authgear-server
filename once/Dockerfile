FROM quay.io/theauthgear/golang:1.23.6-noble AS stage1
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY ./once/ ./
RUN go build -o docker_wrapper -tags 'osusergo netgo static_build timetzdata' .


FROM ubuntu:noble

### A note on apt-get install -y --no-install-recommends --no-install-suggests
###
### We want to make sure we do not install anything that is not essential to running
### the packages we install explicitly.
###
### For example, python3 is a suggested package of postgresql-common.
### But we certainly do not need python3 installed in the image.

## Install less.
## It is useful to view files like postgresql.conf, pg_hba.conf, etc.
RUN set -eux; \
	apt-get update; \
	apt-get install -y --no-install-recommends --no-install-suggests \
		less; \
	rm -rf /var/lib/apt/lists/*

## Install sudo
## We do not run the container as root so we need it so that the user can become root as needed.
RUN set -eux; \
	apt-get update; \
	apt-get install -y --no-install-recommends --no-install-suggests \
		sudo; \
	rm -rf /var/lib/apt/lists/*

## Install locales and set LANG
## initdb uses LANG to determine the locale of the database.
RUN set -eux; \
	apt-get update; \
	apt-get install -y --no-install-recommends --no-install-suggests \
		locales; \
	rm -rf /var/lib/apt/lists/*; \
	echo 'en_US.UTF-8 UTF-8' >> /etc/locale.gen; \
	locale-gen; \
	locale -a | grep 'en_US.utf8'
ENV LANG=en_US.utf8

## Install tzdata
## tzdata is essential to running server that deals with time. (I guess every server deals with time.)
## Note that the version of tzdata is pinned.
RUN set -eux; \
	apt-get update; \
	apt-get install -y --no-install-recommends --no-install-suggests \
		tzdata=2024a-\*; \
	rm -rf /var/lib/apt/lists/*

## Install ca-certificates
## Install root CA certificates.
## This is essential to running programs that connect to external HTTPS server.
RUN set -eux; \
	apt-get update; \
	apt-get install -y --no-install-recommends --no-install-suggests \
		ca-certificates; \
	rm -rf /var/lib/apt/lists/*

## Create the user we use to run the container.
## PostgreSQL does not support running as "root".
## The default user of a PostgreSQL installation is "postgres".
## We do not want to use "postgres" neither.
RUN set -eux; \
	groupadd --system authgear --gid=900; \
	useradd --system --gid=900 --uid=900 --home-dir=/home/authgear --shell=/bin/bash authgear; \
	install --verbose --directory --owner authgear --group authgear --mode=1750 /home/authgear

## Allow the user to run sudo, and run it without password.
RUN set -eux; \
	usermod --append --groups sudo authgear; \
	printf "authgear ALL=(ALL) NOPASSWD:ALL\n" > /etc/sudoers.d/900-authgear

## Install PostgreSQL 16.6
##
##
## We have to install the package postgresql-common first.
## The package postgresql-common installs the following file
##
## /etc/postgresql-common/createcluster.conf
##
## In this file, there is an option create_main_cluster.
## We have to uncomment that option, and set it to false, so that
## when the package postgresql-MAJOR is installed, it does not automatically
## run initdb to create a database that we are not going to use.
##
##
## The installed files belong to the user "postgres".
## We change them back to "authgear".
## \! -path '/proc/*' to skip searching in /proc/ as that would result in file not found error.
##
##
## The following files are installed sample configuration files.
##
## /usr/share/postgresql/16/pg_ident.conf.sample
## /usr/share/postgresql/16/pg_service.conf.sample
## /usr/share/postgresql/16/postgresql.conf.sample
## /usr/share/postgresql/16/pg_hba.conf.sample
##
## In particular, we want to patch the following files:
##
## /usr/share/postgresql/16/postgresql.conf.sample
##
## By default, there is a line `#listen_addresses = 'localhost'`.
## We want to uncomment it and change the value to '*', so that
## the Docker host can access it.
##
## /usr/share/postgresql/16/pg_hba.conf.sample
##
## By default, pg_hba.conf does not specify how to authenticate connection NOT from the loopback address.
## We want to add a line to specify it.
## initdb will replace the token `@authmethodhost@` and `@authmethodlocal@` in this file with the given value of
## --auth-host and --auth-local respectively.
## See https://doxygen.postgresql.org/initdb_8c.html
## So the line we add should use the token.
ENV PG_MAJOR=16
RUN set -eux; \
	apt-get update; \
	apt-get install -y --no-install-recommends --no-install-suggests \
		postgresql-common; \
	sed -E -i "s,^#(create_main_cluster)\\s*=\\s*true,\\1 = false," \
		/usr/share/postgresql-common/createcluster.conf \
		/etc/postgresql-common/createcluster.conf \
	; \
	apt-get install -y --no-install-recommends --no-install-suggests \
		postgresql-16=16.6-\*; \
	rm -rf /var/lib/apt/lists/*; \
	find / \
		-user postgres \
		\! -path '/proc/*' \
		-exec chown authgear:authgear '{}' \; ; \
	sed -E -i "s,^#?(listen_addresses)\\s*=\\s*\\S+,\\1 = '*'," "/usr/share/postgresql/$PG_MAJOR/postgresql.conf.sample"; \
	printf "host\tall\tall\tall\t%s\n" "@authmethodhost@" >> "/usr/share/postgresql/$PG_MAJOR/pg_hba.conf.sample"

ENV PATH=/usr/lib/postgresql/$PG_MAJOR/bin:$PATH

## Install Redis 7.0.x
##
##
## The default configuration of Redis 7.0 can be found here.
## https://raw.githubusercontent.com/redis/redis/7.0/redis.conf
##
## Note that /etc/redis/redis.conf IS NOT the default configuration
## In particular, it has the following changes:
##
## daemonize yes
## pidfile /run/redis/redis-server.pid
## logfile /var/log/redis/redis-server.log
## dir /var/lib/redis
##
## Use this command to what out which option are not commented.
##
##  curl https://raw.githubusercontent.com/redis/redis/7.0/redis.conf | grep -v '^#' | sed -E '/^\s*$/d' > redis.conf
##
## In particular, we need to change the following options.
##
## daemonize no
##   We do not run redis-server as daemon.
##
## pidfile /var/run/redis/redis.pid
##   Write the pid to a location we know.
##
## logfile ""
##   Ask redis-server to write to stdout.
##
## dir /var/lib/redis/data
##   Make sure when the config file is used, the data directory is predictable.
##
## bind * -::*
##   So that the Docker host can access it.
##
## set-proc-title no
##   Keep the original process title.
##
## aclfile
##   Specify password for the default user.
##
## appendonly yes
##   Enable AOF.
##
## auto-aof-rewrite-percentage 50
##   The default is 100%, which means when redis starts with a very large AOF (near the memory limit)
##   it will not be able to rewrite the AOF (since it has no memory to do so).
RUN set -eux; \
	apt-get update; \
	apt-get install -y --no-install-recommends --no-install-suggests \
		redis=5:7.0.\*; \
	rm -rf /var/lib/apt/lists/*; \
	find / \
		-user redis \
		\! -path '/proc/*' \
		-exec chown authgear:authgear '{}' \; ; \
	sed -E -i 's,^(daemonize)\s+.*,\1 no,' /etc/redis/redis.conf; \
	sed -E -i 's,^(pidfile)\s+.*,\1 /run/redis/redis-server.pid,' /etc/redis/redis.conf; \
	sed -E -i 's,^(logfile)\s+.*,\1 "",' /etc/redis/redis.conf; \
	sed -E -i 's,^(dir)\s+.*,\1 /var/lib/redis/data,' /etc/redis/redis.conf; \
	sed -E -i 's,^(bind)\s+.*,\1 * -::*,' /etc/redis/redis.conf; \
	sed -E -i 's,^(set-proc-title)\s+.*,\1 no,' /etc/redis/redis.conf; \
	sed -E -i 's,^#?\s*(aclfile)\s+.*,\1 /var/run/redis/users.acl,' /etc/redis/redis.conf; \
	sed -E -i 's,^(appendonly)\s+.*,\1 yes,' /etc/redis/redis.conf; \
	sed -E -i 's,^(auto-aof-rewrite-percentage)\s+.*,\1 50,' /etc/redis/redis.conf

## Install Ngnix
RUN set -eux; \
	apt-get update; \
	apt-get install -y --no-install-recommends --no-install-suggests \
		nginx; \
	rm -rf /var/lib/apt/lists/*; \
	ln -sf /dev/stdout /var/log/nginx/access.log; \
	ln -sf /dev/stderr /var/log/nginx/error.log
COPY --chown=authgear:authgear ./once/dhparam ./once/nginx.conf.sample /etc/nginx/

## Install certbot.
RUN set -eux; \
	apt-get update; \
	apt-get install -y --no-install-recommends --no-install-suggests \
		certbot \
		python3-certbot-nginx; \
	rm -rf /var/lib/apt/lists/*; \
	cp /etc/letsencrypt/cli.ini /home/authgear/certbot.ini; \
	chown authgear:authgear /home/authgear/certbot.ini


COPY --chown=authgear:authgear ./once/docker-entrypoint.sh ./once/docker-certbot.py /usr/local/bin/
COPY --from=stage1 --chown=authgear:authgear /src/docker_wrapper /usr/local/bin/

ENTRYPOINT ["docker-entrypoint.sh"]

USER authgear
WORKDIR /home/authgear

ENV PGDATA=/var/lib/postgresql/data

VOLUME /var/lib/postgresql/data
VOLUME /var/lib/redis/data
VOLUME /etc/letsencrypt

EXPOSE 5432
EXPOSE 6379
EXPOSE 80
EXPOSE 443

CMD ["docker_wrapper"]
