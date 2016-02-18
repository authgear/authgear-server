#!/bin/sh

go get github.com/tools/godep
go get golang.org/x/tools/cmd/stringer
godep restore
go-wrapper download
go-wrapper install
go generate ./...

mkdir -p dist

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
