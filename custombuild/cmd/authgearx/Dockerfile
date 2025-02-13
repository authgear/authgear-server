# syntax=docker/dockerfile:1

# Stage 0: Build the static files
FROM node:20.9.0-bookworm AS authgear-stage-0
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
FROM quay.io/theauthgear/golang:1.23.6-noble AS authgear-stage-1

# Install build time C dependencies
RUN apt-get update && apt-get install -y --no-install-recommends --no-install-suggests \
    pkg-config \
    libicu-dev \
    libvips-dev \
    libmagic-dev \
    && rm -rf /var/lib/apt/lists/*

# In order to build a Go program that uses private modules in Docker,
# we need the following
#
# 1. Set GOPRIVATE
# 2. Set up ~/.gitconfig to make Go to use SSH instead HTTPS to fetch the private modules.
# 3. Set up ~/.ssh/known_hosts
# 4. use --mount=type=ssh to use the SSH agent from the host machine.
ENV GOPRIVATE github.com/authgear/iamsmart
RUN git config --global url."ssh://git@github.com/authgear/iamsmart".insteadOf https://github.com/authgear/iamsmart
RUN mkdir -p ~/.ssh \
  && ssh-keyscan github.com >> ~/.ssh/known_hosts

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
WORKDIR /src/custombuild
COPY ./custombuild/go.mod ./custombuild/go.sum ./
RUN --mount=type=ssh go mod download
WORKDIR /src
COPY . .
# Let //go:embed to embed the built static files
COPY --from=authgear-stage-0 /usr/src/app/resources/authgear/ ./resources/authgear/
ARG GIT_HASH
RUN make -C custombuild build BIN_NAME=authgear TARGET=authgearx GIT_HASH=$GIT_HASH

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
            tzdata; \
        dpkg --status tzdata | awk '$1 ~ /^Version:$/ { print $2 }' | grep '^2024b'; \
        rm -rf /var/lib/apt/lists/*
COPY --from=authgear-stage-1 /src/custombuild/authgear /usr/local/bin/
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
EXPOSE 3000
CMD ["authgear", "start"]
