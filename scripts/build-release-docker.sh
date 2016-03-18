#!/bin/sh

: ${SKYGEAR_VERSION:=latest}
IMAGE_NAME=skygeario/skygear-server:v$SKYGEAR_VERSION

curl --fail -sIL "https://dl.bintray.com/skygeario/skygear/skygear-server/v$SKYGEAR_VERSION/skygear-server-v$SKYGEAR_VERSION-zmq-linux-amd64" > /dev/null

if [ $? -eq 0 ]; then
    docker build --pull -t $IMAGE_NAME --build-arg SKYGEAR_VERSION=$SKYGEAR_VERSION -f Dockerfile-release .
    echo "Done. Run \`docker push $IMAGE_NAME\` to push image."
else
    echo "Skygear(v$SKYGEAR_VERSION) is not available on Bintray."
fi
