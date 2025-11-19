# syntax=docker/dockerfile:1

# Stage 0: Build the static files
FROM node:20.19.5-bookworm AS authgear-portal-stage-0
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

# Stage 1: Build the portal static files
FROM node:20.19.5-bookworm AS authgear-portal-stage-1
# If the working directory is /src, Parcel will have some problem with it.
WORKDIR /usr/src/app
COPY ./portal/package.json ./portal/package-lock.json ./
# Copy the scripts
COPY ./portal/scripts/. ./scripts/.
RUN npm ci
COPY ./portal .
RUN npm run build

# Stage 2: Build the Go binary
FROM quay.io/theauthgear/golang:1.25.4-noble AS authgear-portal-stage-2

# Install build time C dependencies
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends --no-install-suggests \
        pkg-config \
        libicu-dev \
        libvips-dev \
        libmagic-dev; \
    rm -rf /var/lib/apt/lists/*

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
# Only copy the necessary files to improving caching.
COPY ./Makefile ./embed.go ./
COPY ./makefiles/ ./makefiles/
COPY ./resources/ ./resources/
COPY ./cmd/ ./cmd/
COPY ./pkg/ ./pkg/
COPY ./custombuild/ ./custombuild/
# Let //go:embed to embed the built static files
COPY --from=authgear-portal-stage-0 /usr/src/app/resources/authgear/ ./resources/authgear/
COPY --from=authgear-portal-stage-1 /usr/src/app/dist/ ./resources/portal/static/
ARG GIT_HASH
ARG AUTHGEARONCE
ARG AUTHGEARLITE
RUN make -C custombuild build BIN_NAME=authgear-portal TARGET=portalx

# We used to build static binary.
# But we have a transitive dependency on icu4c so this is no longer the case.
# RUN readelf -d ./authgear | grep 'There is no dynamic section in this file'

# Runtime stage: Prepare the runtime dependencies
FROM ubuntu:noble AS authgear-portal-stage-runtime
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
COPY --from=authgear-portal-stage-2 /src/custombuild/authgear-portal /usr/local/bin/
# There is no need to run update-ca-certificates
# because what it does is install extra certificates.
# Those certificates typically are mounted at runtime.
# So at build time (i.e. here and now), there are no extra certificates to be installed.
# RUN update-ca-certificates

# Final Stage: Define WORKDIR, ENTRYPOINT, and CMD
FROM authgear-portal-stage-runtime AS authgear-portal-stage-final
WORKDIR /app
COPY ./docker-entrypoint.sh /usr/local/bin/
ENTRYPOINT ["docker-entrypoint.sh"]
# update-ca-certificates requires root to run.
#USER nobody

# portal
EXPOSE 3003
# portal internal
EXPOSE 13003

CMD ["authgear-portal", "start"]
