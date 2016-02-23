FROM golang:1.5

RUN \
    apt-get update && \
    apt-get install --no-install-recommends -y libtool-bin automake pkg-config && \
    git clone --branch 1.0.5 --depth 1 git://github.com/jedisct1/libsodium.git && \
    ( cd libsodium; ./autogen.sh; ./configure; make install; ldconfig ) && \
    rm -rf libsodium && \
    git clone --branch v4.1.3 --depth 1 git://github.com/zeromq/zeromq4-1.git && \
    ( cd zeromq4-1; ./autogen.sh; ./configure; make install; ldconfig ) && \
    rm -rf zeromq4-1 && \
    git clone --branch v3.0.2 --depth 1 git://github.com/zeromq/czmq.git && \
    ( cd czmq; ./autogen.sh; ./configure; make install; ldconfig ) && \
    rm -rf czmq && \
    apt-get install --no-install-recommends -y net-tools telnet && \
    rm -rf /var/lib/apt/lists/*

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

# Copy a minimal set of files to restore Go dependencies to get advantage
# of Docker build cache
RUN go get github.com/tools/godep
COPY Godeps /go/src/app/Godeps
RUN $GOPATH/bin/godep restore

COPY . /go/src/app

RUN go-wrapper download && \
    go-wrapper install --tags zmq

VOLUME /go/src/app/data

EXPOSE 3000

CMD ["go-wrapper", "run"]
