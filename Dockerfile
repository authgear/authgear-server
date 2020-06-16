FROM golang:1.14.4-buster as build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make build
# Check if the binary is really static
RUN readelf -d ./authgear | grep 'There is no dynamic section in this file'

FROM debian:buster-slim
WORKDIR /app
# /etc/ssl/certs
# /etc/mime.types
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    mime-support \
    && rm -rf /var/lib/apt/lists/*
RUN update-ca-certificates
COPY ./reserved_name.txt .
COPY --from=build /src/authgear .
USER nobody
EXPOSE 3000
CMD ["/app/authgear"]
