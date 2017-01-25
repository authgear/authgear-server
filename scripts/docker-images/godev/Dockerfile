FROM golang:1.7

ENV LIBSODIUM_VERSION=1.0.5 \
    ZMQ_VERSION=4.1.3 \
    CZMQ_VERSION=3.0.2

RUN \
    apt-get update && \
    apt-get install --no-install-recommends -y libtool-bin automake pkg-config && \
    git clone --branch $LIBSODIUM_VERSION --depth 1 git://github.com/jedisct1/libsodium.git && \
    ( cd libsodium; ./autogen.sh; ./configure; make check; make install; ldconfig ) && \
    rm -rf libsodium && \
    git clone --branch v$ZMQ_VERSION --depth 1 git://github.com/zeromq/zeromq4-1.git && \
    ( cd zeromq4-1; ./autogen.sh; ./configure; make check; make install; ldconfig ) && \
    rm -rf zeromq4-1 && \
    git clone --branch v$CZMQ_VERSION --depth 1 git://github.com/zeromq/czmq.git && \
    ( cd czmq; ./autogen.sh; ./configure; make check; make install; ldconfig ) && \
    rm -rf czmq && \
    rm -rf /var/lib/apt/lists/*

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

RUN go get github.com/Masterminds/glide \
           golang.org/x/tools/cmd/stringer \
           golang.org/x/tools/cmd/cover \
           github.com/golang/lint/golint \
           github.com/golang/mock/mockgen \
           github.com/rickmak/gocyclo \
           github.com/oursky/gogocyclo \
           github.com/mitchellh/gox \
           github.com/vektra/mockery/.../
