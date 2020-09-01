# Stage 1: Build the Go binary
FROM golang:1.15.0-buster as stage1
ARG GIT_HASH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make build BIN_NAME=authgear-portal TARGET=portal GIT_HASH=$GIT_HASH
# Check if the binary is really static
RUN readelf -d ./authgear-portal | grep 'There is no dynamic section in this file'

# Stage 2: build the static files
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
    ca-certificates \
    mime-support \
    tzdata \
    && rm -rf /var/lib/apt/lists/*
RUN update-ca-certificates
COPY --from=stage1 /src/authgear-portal /usr/local/bin/
# FIXME(portal): Serve static asset
USER nobody
EXPOSE 3003
CMD ["authgear-portal", "start"]
