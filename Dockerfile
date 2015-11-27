FROM golang:1.5.1

RUN \
    apt-get update && \
    apt-get install --no-install-recommends -y libtool-bin automake pkg-config libsodium-dev libzmq3-dev && \
    git clone --branch v3.0.2 --depth 1 git://github.com/zeromq/czmq.git && \
    ( cd czmq; ./autogen.sh; ./configure; make check; make install; ldconfig ) && \
    rm -rf czmq && \
    rm -rf /var/lib/apt/lists/*

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

# Copy a minimal set of files to restore Go dependencies to get advantage
# of Docker build cache
RUN go get github.com/tools/godep
COPY Godeps /go/src/app/Godeps
RUN $GOPATH/bin/godep restore

COPY . /go/src/app

RUN go-wrapper download
RUN go-wrapper install

VOLUME /go/src/app/data

EXPOSE 3000

CMD ["go-wrapper", "run"]
