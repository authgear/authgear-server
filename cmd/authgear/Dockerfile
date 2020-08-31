# Stage 1: Build the Go binary
FROM golang:1.14.4-buster as stage1
ARG GIT_HASH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make build GIT_HASH=$GIT_HASH
# Check if the binary is really static
RUN readelf -d ./authgear | grep 'There is no dynamic section in this file'

# Stage 2: Build the static files
FROM node:12.18.0-buster as stage2
ARG GIT_HASH
WORKDIR /src
COPY ./scripts/npm/package.json ./scripts/npm/package-lock.json ./scripts/npm/
RUN cd ./scripts/npm && npm ci
COPY . .
RUN make static GIT_HASH=$GIT_HASH
RUN make html-email && rm ./templates/*.mjml

FROM debian:buster-slim
ARG GIT_HASH
WORKDIR /app
# /etc/ssl/certs
# /etc/mime.types
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    mime-support \
    && rm -rf /var/lib/apt/lists/*
RUN update-ca-certificates
COPY ./reserved_name.txt .
COPY ./migrations ./migrations
COPY --from=stage1 /src/authgear /usr/local/bin/
COPY --from=stage2 /src/dist/ ./dist/
COPY --from=stage2 /src/templates/ ./templates/
USER nobody
EXPOSE 3000
ENV STATIC_ASSET_DIR ./dist
ENV STATIC_ASSET_URL_PREFIX /static/$GIT_HASH
CMD ["authgear", "start"]
