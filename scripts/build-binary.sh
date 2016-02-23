#!/bin/sh

go get github.com/tools/godep
go get golang.org/x/tools/cmd/stringer
godep restore
go-wrapper download
go-wrapper install
go generate ./...

mkdir -p dist

# build skygear server without C bindings
for GOOS in darwin linux windows; do
  for GOARCH in 386 amd64; do
    if [ -n "$VERSION" ]; then
        FILENAME=skygear-$VERSION-$GOOS-$GOARCH
   else
        FILENAME=skygear-$GOOS-$GOARCH
    fi
    GOOS=$GOOS GOARCH=$GOARCH go build -o dist/$FILENAME github.com/oursky/skygear
  done
done

# build skygear server with zmq
ldconfig -p | grep libczma
if [ $? -eq 0 ]; then
  GOOS=linux
  GOARCH=amd64
  if [ -n "$VERSION" ]; then
    FILENAME=skygear-zmq-$VERSION-$GOOS-$GOARCH
  else
    FILENAME=skygear-zmq-$GOOS-$GOARCH
  fi
  GOOS=$GOOS GOARCH=$GOARCH go build --tags zmq -o dist/$FILENAME github.com/oursky/skygear
else
  >&2 echo "Did not build skygear with zmq because libczmq library is not installed."
fi
