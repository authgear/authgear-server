# Stage 1: Build the Go binary
FROM golang:1.15.0-buster as stage1

# Install build time C dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    pkg-config \
    libicu-dev \
    && rm -rf /var/lib/apt/lists/*

ARG GIT_HASH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make build BIN_NAME=authgear-portal TARGET=portal GIT_HASH=$GIT_HASH

# We used to build static binary.
# But we have a transitive dependency on icu4c so this is no longer the case.
# RUN readelf -d ./authgear | grep 'There is no dynamic section in this file'

# Stage 2: Build the portal static files
FROM node:12.18.3-buster as stage2
ARG GIT_HASH
# If the working directory is /src, Parcel will have some problem with it.
WORKDIR /usr/src/app
COPY ./portal/package.json ./portal/package-lock.json ./
RUN npm ci
COPY ./portal .
RUN npm run build

# Stage 3: Prepare the actual fs we use to run the program
FROM debian:buster-slim
ARG GIT_HASH
WORKDIR /app
# /etc/mime.types (mime-support)
# /usr/share/ca-certificates/*/* (ca-certificates)
# /usr/share/zoneinfo/ (tzdata)
RUN apt-get update && apt-get install -y --no-install-recommends \
    libicu-dev \
    ca-certificates \
    mime-support \
    tzdata \
    && rm -rf /var/lib/apt/lists/*
RUN update-ca-certificates
COPY --from=stage1 /src/authgear-portal /usr/local/bin/
COPY --from=stage2 /usr/src/app/dist/ ./static/
COPY ./templates/ ./templates/
ENV STATIC_ASSET_SERVING_ENABLED true
USER nobody
EXPOSE 3003
CMD ["authgear-portal", "start"]
