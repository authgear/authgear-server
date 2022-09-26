#!/bin/sh

set -eu
# Make the certificates in /usr/local/share/ca-certificates take effect.
update-ca-certificates
# Run the CMD specified in Dockerfile, or the CMD specified by the user.
exec "$@"
