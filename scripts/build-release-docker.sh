#!/bin/sh

: ${SKYGEAR_VERSION:=latest}

# Run this script in the project root directory.

docker build -t skygear-build -f Dockerfile-development .
docker run -it --rm -v `pwd`:/go/src/app -w /go/src/app -e GOOS=linux -e GOARCH=amd64 skygear-build go build --tags zmq -o skygear-server github.com/oursky/skygear
docker build -t skygeario/skygear-server:$SKYGEAR_VERSION -f Dockerfile-release .
docker push skygeario/skygear-server:$SKYGEAR_VERSION


