#!/bin/sh

if [ -z "$GITHUB_REF_TYPE" ]; then
	echo 1>&2 "GITHUB_REF_TYPE is not set"
	exit 1
fi
if [ -z "$GITHUB_REF_NAME" ]; then
	echo 1>&2 "GITHUB_REF_NAME is not set"
	exit 1
fi

if [ "$GITHUB_REF_TYPE" = "tag" ]; then
	if printf "%s" "$GITHUB_REF_NAME" | grep 'alpha' 1>/dev/null 2>/dev/null; then
		printf "AUTHGEARONCE_LICENSE_SERVER_ENV=staging\n"
	else
		printf "AUTHGEARONCE_LICENSE_SERVER_ENV=production\n"
	fi
fi
