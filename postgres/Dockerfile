FROM postgres:12.3

ENV PARTMAN_VERSION 4.5.1

RUN apt-get update && apt-get install -y \
	unzip \
	build-essential \
	postgresql-server-dev-11 \
	wget \
	&& rm -rf /var/lib/apt/lists/*

RUN wget https://github.com/pgpartman/pg_partman/archive/v${PARTMAN_VERSION}.zip -O pg_partman-${PARTMAN_VERSION}.zip && unzip pg_partman-${PARTMAN_VERSION}.zip && cd pg_partman-${PARTMAN_VERSION} && make NO_BGW=1 install
