#!/bin/sh

set -e

: ${SKYGEAR_VERSION:=latest}
IMAGE_NAME=skygeario/skygear-server:v$SKYGEAR_VERSION

if [ -d dist ]; then
    echo "Error: Directory 'dist' exists."
    exit 1
fi

if [ -f skygear-server ]; then
    echo "Error: File 'skygear-server' exists."
    exit 1
fi

docker build -t skygear-build -f Dockerfile-development .
docker run -it --rm -v `pwd`:/go/src/app -w /go/src/app skygear-build /go/src/app/scripts/build-binary.sh
cp dist/skygear-server-zmq-linux-amd64 skygear-server
docker build --pull -t $IMAGE_NAME -f Dockerfile-release .

echo "Done. Run \`docker push $IMAGE_NAME\` to push image."
