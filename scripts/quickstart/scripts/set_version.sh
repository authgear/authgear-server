#!/bin/sh

set -e

: ${DIST:=/home/ubuntu/myapp}
: ${SKYGEAR_VERSION:=latest}

sed -i -e "s/image: quay.io\/skygeario\/skygear-server:latest/image: quay.io\/skygeario\/skygear-server:${SKYGEAR_VERSION}/" $DIST/docker-compose.yml
