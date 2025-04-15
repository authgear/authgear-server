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
	version="$(printf "%s" "$GITHUB_REF_NAME" | sed -n 's,^authgear-once/\(.*\)$,\1,p')"
	if [ -z "$version" ]; then
		echo 1>&2 "bad version: $GITHUB_REF_NAME"
		exit 1
	fi
	printf "GIT_TAG_NAME=%s\n" "$version"
fi
