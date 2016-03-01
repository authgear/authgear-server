#!/bin/sh

set -e

: ${DIST:=/home/ubuntu/myapp}

docker-compose -f ~/myapp/docker-compose.yml pull



