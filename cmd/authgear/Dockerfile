# syntax=docker/dockerfile:1

# Stage 0: Build the static files
FROM --platform=$BUILDPLATFORM node:20.9.0-bookworm AS authgear-stage-0
WORKDIR /usr/src/app
COPY ./authui/package.json ./authui/package-lock.json ./authui/
# See /authui/README.md for the reason why we need to copy tarballs.
COPY ./authui/tarballs/. ./authui/tarballs/.
RUN cd ./authui && npm ci
# Only copy the necessary files to improving caching.
COPY ./resources/ ./resources/
COPY ./authui/ ./authui/
COPY ./makefiles/ ./makefiles/
COPY Makefile .
RUN make authui

# Stage 1: Build the Go binary
FROM quay.io/theauthgear/golang:1.25.4-noble AS authgear-stage-1

# Install build time C dependencies
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends --no-install-suggests \
        pkg-config \
        libicu-dev \
        libvips-dev \
        libmagic-dev; \
    rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
# Only copy the necessary files to improving caching.
COPY ./Makefile ./embed.go ./
COPY ./makefiles/ ./makefiles/
COPY ./resources/ ./resources/
COPY ./cmd/ ./cmd/
COPY ./pkg/ ./pkg/
COPY ./custombuild/ ./custombuild/
# Let //go:embed to embed the built static files
COPY --from=authgear-stage-0 /usr/src/app/resources/authgear/ ./resources/authgear/
ARG GIT_HASH
ARG AUTHGEARONCE
ARG AUTHGEARLITE
RUN make build BIN_NAME=authgear TARGET=authgear

# We used to build static binary.
# But we have a transitive dependency on icu4c so this is no longer the case.
# RUN readelf -d ./authgear | grep 'There is no dynamic section in this file'

# Runtime stage: Prepare the runtime dependencies
FROM ubuntu:noble AS authgear-stage-runtime
# /etc/mime.types (media-types)
# /usr/share/ca-certificates/*/* (ca-certificates)
# /usr/share/zoneinfo/ (tzdata)
RUN set -eux; \
        apt-get update; \
        apt-get install -y --no-install-recommends --no-install-suggests \
            libicu-dev \
            libvips-dev \
            libmagic-dev \
            libmagic-mgc \
            ca-certificates \
            media-types \
            tzdata=2025b\*; \
        rm -rf /var/lib/apt/lists/*
COPY --from=authgear-stage-1 /src/authgear /usr/local/bin/
# There is no need to run update-ca-certificates
# because what it does is install extra certificates.
# Those certificates typically are mounted at runtime.
# So at build time (i.e. here and now), there are no extra certificates to be installed.
# RUN update-ca-certificates

# Final Stage: Define WORKDIR, ENTRYPOINT, and CMD
FROM authgear-stage-runtime AS authgear-stage-final
WORKDIR /app
COPY ./docker-entrypoint.sh /usr/local/bin/
ENTRYPOINT ["docker-entrypoint.sh"]
# update-ca-certificates requires root to run.
#USER nobody

# main
EXPOSE 3000
# main internal
EXPOSE 13000
# resolver
EXPOSE 3001
# resolver internal
EXPOSE 13001
# admin
EXPOSE 3002
# admin internal
EXPOSE 13002

CMD ["authgear", "start"]
