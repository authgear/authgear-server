#!/bin/sh

set -e

if [ ! -d dist ]; then
    echo "Error: Directory 'dist' does not exist."
    exit 1
fi

SUBJECT_NAME="skygeario-builder"
ORG_NAME="skygeario"
PACKAGE_NAME="skygear-server"
VERSION=`git describe --tags`

if [ -n "$TRAVIS_TAG" ]; then
    REPO_NAME="skygear"
    PACKAGE_VERSION="$TRAVIS_TAG"
else
    REPO_NAME="skygear-nightly"
    PACKAGE_VERSION=`git rev-parse --abbrev-ref HEAD`
fi

for PER_FILE in `ls -1 dist | sed s/skygear-server-/skygear-server-$VERSION-/`; do
    UPLOAD_URL="https://api.bintray.com/content/$ORG_NAME/$REPO_NAME/$PACKAGE_NAME/$PACKAGE_VERSION/$PER_FILE"

    echo "\nUploading \"$PER_FILE\" to \"$UPLOAD_URL\"..."

    curl -T "dist/$PER_FILE" \
    -H "X-Bintray-Publish: 1" \
    -H "X-Bintray-Override: 1" \
    -u$SUBJECT_NAME:$BINTRAY_API_KEY \
    $UPLOAD_URL
done
