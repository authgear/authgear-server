FROM golang:1.4.2

RUN \
    apt-get update && \
    apt-get install --no-install-recommends -y libtool automake pkg-config libsodium-dev libzmq3-dev && \
    git clone git://github.com/zeromq/czmq.git && \
    ( cd czmq; ./autogen.sh; ./configure; make check; make install; ldconfig ) && \
    rm -rf czmq && \
    rm -rf /var/lib/apt/lists/*

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

COPY . /go/src/app

RUN go get github.com/tools/godep
RUN $GOPATH/bin/godep restore

RUN go-wrapper download
RUN go-wrapper install

RUN \
    go get golang.org/x/tools/cmd/cover && \
    go get github.com/golang/lint/golint && \
    go get github.com/smartystreets/goconvey/convey && \
    go get github.com/smartystreets/assertions && \
    go get golang.org/x/tools/cmd/stringer && \
    golint ./... && \
    go generate ./...

EXPOSE 3000

CMD ["go-wrapper", "run"]
