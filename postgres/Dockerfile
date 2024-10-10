FROM postgres:12.20-bookworm

# Google Cloud SQL PostgreSQL 12 supports up to 4.7.4.
# https://cloud.google.com/sql/docs/postgres/extensions#:~:text=Cloud%20SQL%20for%20PostgreSQL%20versions%2014%20and%20later%20support%20version%205.0.1%20while%20Cloud%20SQL%20for%20PostgreSQL%20versions%20that%20are%20earlier%20than%2014%20support%20only%20up%20to%20version%204.7.4.
# So we use that version here.
ENV PARTMAN_VERSION 4.7.4

RUN apt-get update && apt-get install -y \
	unzip \
	build-essential \
	postgresql-server-dev-all \
	wget \
	&& rm -rf /var/lib/apt/lists/*

RUN wget https://github.com/pgpartman/pg_partman/archive/v${PARTMAN_VERSION}.zip -O pg_partman-${PARTMAN_VERSION}.zip && unzip pg_partman-${PARTMAN_VERSION}.zip && cd pg_partman-${PARTMAN_VERSION} && make NO_BGW=1 install
