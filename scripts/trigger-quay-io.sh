#!/bin/bash -e

SUBJECT_NAME="skygeario-builder"
ORG_NAME="skygeario"
PACKAGE_NAME="skygear-server"
VERSION=`git describe --tags`

if [ -n "$TRAVIS_TAG" ]; then
    REPO_NAME="skygear"
    PACKAGE_VERSION="$TRAVIS_TAG"
else
    REPO_NAME="skygear-nightly"
    PACKAGE_VERSION=$TRAVIS_BRANCH
fi

DOCKER_BUILD_ARCHIVE="docker-build-archive-$VERSION.tar.gz"
TEMP_BUILD_DIR=`mktemp -d`
ls -al dist
cp dist/skygear-server-zmq-linux-amd64 $TEMP_BUILD_DIR/skygear-server
cp scripts/docker-release/Dockerfile $TEMP_BUILD_DIR
cat >> $TEMP_BUILD_DIR/Dockerfile << EOF
ENV SKYGEAR_VERSION=$VERSION
LABEL \
    io.skygear.role=server \
    io.skygear.repo=SkygearIO/skygear-server \
    io.skygear.commit=`git rev-parse HEAD` \
    io.skygear.version=$SKYGEAR_VERSION \
    io.skygear.build-date=`date -u +"%Y-%m-%dT%H:%M:%SZ"`
EOF
tar -C $TEMP_BUILD_DIR -zcvf dist/$DOCKER_BUILD_ARCHIVE .
rm -rf $TEMP_BUILD_DIR

UPLOAD_URL="https://api.bintray.com/content/$ORG_NAME/$REPO_NAME/$PACKAGE_NAME/$PACKAGE_VERSION/$DOCKER_BUILD_ARCHIVE"

>&2 echo "\nUploading \"$DOCKER_BUILD_ARCHIVE\" to \"$UPLOAD_URL\"..."
curl -T "dist/$DOCKER_BUILD_ARCHIVE" \
    -H "X-Bintray-Publish: 1" \
    -H "X-Bintray-Override: 1" \
    -u$SUBJECT_NAME:$BINTRAY_API_KEY \
    $UPLOAD_URL

QUAY_IO_TRIGGER_URL=https://quay.io/api/v1/repository/$ORG_NAME/$PACKAGE_NAME/build/
QUAY_IO_ARCHIVE_URL=https://bintray.com/artifact/download/$ORG_NAME/$REPO_NAME/$DOCKER_BUILD_ARCHIVE

IMAGE_TAG_NAME=${PACKAGE_VERSION/master/canary}

>&2 echo "Trigger build on Quay.io..."
curl -H "Content-Type: application/json" \
    -H "Authorization: Bearer $QUAY_IO_ACCESS_TOKEN" \
    --data '{"docker_tags": ["'$IMAGE_TAG_NAME'", "git-'`git rev-parse --short HEAD`'"], "archive_url": "'$QUAY_IO_ARCHIVE_URL'"}' \
    -X POST \
    $QUAY_IO_TRIGGER_URL

