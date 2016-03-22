#!/bin/sh

set -e

DAEMON_NAME=skygear-server
VERSION=`git describe --always --tags`

mkdir -p dist

# build skygear server without C bindings
for GOOS in darwin linux windows; do
  for GOARCH in 386 amd64; do
    FILENAME=$DAEMON_NAME-$GOOS-$GOARCH
    echo -n "Building $FILENAME... "

    GOOS=$GOOS GOARCH=$GOARCH \
    go build \
    -ldflags "-X github.com/skygeario/skygear-server/skyversion.version=$VERSION" \
    -o dist/$FILENAME \
    github.com/skygeario/skygear-server

    echo "Done"
  done
done

# build skygear server with zmq
GOOS=linux
GOARCH=amd64
FILENAME=$DAEMON_NAME-zmq-$GOOS-$GOARCH
echo -n "Building $FILENAME... "

GOOS=$GOOS GOARCH=$GOARCH \
go build \
--tags zmq \
-ldflags "-X github.com/skygeario/skygear-server/skyversion.version=$VERSION" \
-o dist/$FILENAME \
github.com/skygeario/skygear-server

echo "Done"
