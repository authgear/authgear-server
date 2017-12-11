FROM golang:1.8.3-stretch as godev

RUN \
    apt-get update && \
    apt-get install --no-install-recommends -y libtool-bin automake pkg-config libsodium-dev libzmq3-dev && \
    rm -rf /var/lib/apt/lists/* && \
    go get github.com/golang/dep/cmd/dep

RUN mkdir -p /go/src/github.com/skygeario/skygear-server
WORKDIR /go/src/github.com/skygeario/skygear-server

# Copy a minimal set of files to restore Go dependencies to get advantage
# of Docker build cache
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only

ARG VERSION=

COPY . .
RUN make build VERSION=$VERSION

FROM alpine:3.4

RUN apk --update --no-cache add libc6-compat libstdc++ zlib ca-certificates \
        libsodium libzmq && \
    ln -s /lib /lib64

COPY --from=godev /go/src/github.com/skygeario/skygear-server/skygear-server /usr/local/bin/
RUN mkdir -p /app/data \
    && chown nobody:nobody /app/data

WORKDIR /app
VOLUME /app/data
USER nobody

EXPOSE 3000

CMD ["/usr/local/bin/skygear-server"]
