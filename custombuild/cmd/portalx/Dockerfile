# syntax=docker/dockerfile:1

# Stage 1: Build the Go binary
FROM golang:1.22.4-bookworm as stage1

# Install build time C dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
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
ARG GIT_HASH
RUN make -C custombuild build BIN_NAME=authgear-portal TARGET=portalx GIT_HASH=$GIT_HASH

# We used to build static binary.
# But we have a transitive dependency on icu4c so this is no longer the case.
# RUN readelf -d ./authgear | grep 'There is no dynamic section in this file'

# Stage 2: Build the static files
FROM node:20.9.0-bookworm as stage2
ARG GIT_HASH
WORKDIR /usr/src/app
COPY ./scripts/npm/package.json ./scripts/npm/package-lock.json ./scripts/npm/
RUN cd ./scripts/npm && npm ci
COPY ./authui/package.json ./authui/package-lock.json ./authui/
RUN cd ./authui && npm ci
COPY . .
RUN make authui GIT_HASH=$GIT_HASH

# Stage 3: Build the portal static files
FROM node:20.9.0-bookworm as stage3
ARG GIT_HASH
# If the working directory is /src, Parcel will have some problem with it.
WORKDIR /usr/src/app
COPY ./portal/package.json ./portal/package-lock.json ./
RUN npm ci
COPY ./portal .
RUN npm run build

# Stage 4: Prepare the actual fs we use to run the program
FROM debian:bookworm-slim
ARG GIT_HASH
WORKDIR /app
# /etc/mime.types (mime-support)
# /usr/share/ca-certificates/*/* (ca-certificates)
# /usr/share/zoneinfo/ (tzdata)
RUN apt-get update && apt-get install -y --no-install-recommends \
    libicu-dev \
    libvips-dev \
    libmagic-dev \
    libmagic-mgc \
    ca-certificates \
    mime-support \
    tzdata \
    && rm -rf /var/lib/apt/lists/*
RUN update-ca-certificates
COPY ./GeoLite2-Country.mmdb ./GeoLite2-Country.mmdb
COPY ./migrations ./migrations
COPY --from=stage1 /src/custombuild/authgear-portal /usr/local/bin/
COPY ./resources/ ./resources/
COPY --from=stage2 /usr/src/app/resources/authgear/ ./resources/authgear/
COPY --from=stage3 /usr/src/app/dist/ ./resources/portal/static/
COPY ./docker-entrypoint.sh /usr/local/bin/
ENTRYPOINT ["docker-entrypoint.sh"]
# update-ca-certificates requires root to run.
#USER nobody
EXPOSE 3003
CMD ["authgear-portal", "start"]
