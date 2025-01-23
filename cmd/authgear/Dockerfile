# syntax=docker/dockerfile:1

# Stage 0: Build the static files
FROM --platform=$BUILDPLATFORM node:20.9.0-bookworm AS stage0
ARG GIT_HASH
WORKDIR /usr/src/app
COPY ./scripts/npm/package.json ./scripts/npm/package-lock.json ./scripts/npm/
RUN cd ./scripts/npm && npm ci
COPY ./authui/package.json ./authui/package-lock.json ./authui/
# See /authui/README.md for the reason why we need to copy tarballs.
COPY ./authui/tarballs/. ./authui/tarballs/.
RUN cd ./authui && npm ci
COPY . .
RUN make authui GIT_HASH=$GIT_HASH

# Stage 1: Build the Go binary
FROM quay.io/theauthgear/golang:1.23.4-noble AS stage1

# Install build time C dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    pkg-config \
    libicu-dev \
    libvips-dev \
    libmagic-dev \
    && rm -rf /var/lib/apt/lists/*

ARG GIT_HASH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Let //go:embed to embed the built static files
COPY --from=stage0 /usr/src/app/resources/authgear/ ./resources/authgear/
RUN make build BIN_NAME=authgear TARGET=authgear GIT_HASH=$GIT_HASH

# We used to build static binary.
# But we have a transitive dependency on icu4c so this is no longer the case.
# RUN readelf -d ./authgear | grep 'There is no dynamic section in this file'

# Final Stage: Prepare the actual fs we use to run the program
FROM ubuntu:noble
ARG GIT_HASH
WORKDIR /app
# /etc/mime.types (media-types)
# /usr/share/ca-certificates/*/* (ca-certificates)
# /usr/share/zoneinfo/ (tzdata)
RUN apt-get update && apt-get install -y --no-install-recommends \
    libicu-dev \
    libvips-dev \
    libmagic-dev \
    libmagic-mgc \
    ca-certificates \
    media-types \
    tzdata \
    && apt-get upgrade -y && rm -rf /var/lib/apt/lists/*
RUN update-ca-certificates
COPY --from=stage1 /src/authgear /usr/local/bin/
COPY ./docker-entrypoint.sh /usr/local/bin/
ENTRYPOINT ["docker-entrypoint.sh"]
# update-ca-certificates requires root to run.
#USER nobody
EXPOSE 3000
CMD ["authgear", "start"]
