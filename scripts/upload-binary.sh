#!/bin/sh

set -e

if [ ! -d dist ]; then
    echo "Error: Directory 'dist' does not exist."
    exit 1
fi

COMMIT_NAME=`git rev-parse --short HEAD`
TIMESTAMP=`date +%s`
SUBJECT_NAME="ben181231"
ORG_NAME="skygeario"
REPO_NAME="skygear"
PACKAGE_NAME="skygear-server"

if [ -n "$TRAVIS_TAG" ]; then
    PACKAGE_VERSION="$TRAVIS_TAG"
    FOLDER_STRUCTURE="$PACKAGE_NAME/Releases/$TRAVIS_TAG"
else
    PACKAGE_VERSION="Commits-$TIMESTAMP-$COMMIT_NAME"
    FOLDER_STRUCTURE="$PACKAGE_NAME/Commits/$TIMESTAMP-$COMMIT_NAME"
fi

for PER_FILE in `ls -1 dist`; do
    UPLOAD_URL="https://api.bintray.com/content/$ORG_NAME/$REPO_NAME/$PACKAGE_NAME/$PACKAGE_VERSION/$FOLDER_STRUCTURE/$PER_FILE"

    echo "\nUploading \"$PER_FILE\" to \"$UPLOAD_URL\"..."

    curl -T "dist/$PER_FILE" \
    -H "X-Bintray-Publish: 1" \
    -H "X-Bintray-Override: 1" \
    -u$SUBJECT_NAME:$BINTRAY_API_KEY \
    $UPLOAD_URL
done
