#!/bin/sh

set -e

export PGUSER=postgres

cd /usr/share/postgresql/$PG_MAJOR/contrib/postgis-$POSTGIS_MAJOR
psql --dbname postgres --command 'CREATE EXTENSION postgis;'
