#!/bin/sh

set -e

DAEMON_NAME=skygear-server
DIST=dist

mkdir -p $DIST

# build skygear server without C bindings
for GOOS in darwin linux windows; do
  for GOARCH in 386 amd64; do
    FILENAME=$DAEMON_NAME-$GOOS-$GOARCH
    GOOS=$GOOS GOARCH=$GOARCH go build -o $DIST/$FILENAME github.com/oursky/skygear
  done
done

# build skygear server with zmq
ldconfig -p | grep libczmq
if [ $? -eq 0 ]; then
  GOOS=linux
  GOARCH=amd64
  FILENAME=$DAEMON_NAME-zmq-$GOOS-$GOARCH
  GOOS=$GOOS GOARCH=$GOARCH go build --tags zmq -o $DIST/$FILENAME github.com/oursky/skygear
else
  >&2 echo "Did not build skygear with zmq because libczmq library is not installed."
fi
