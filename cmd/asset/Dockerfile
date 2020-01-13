FROM golang:1.13.6-alpine3.11 as godev

# alpine-sdk includes git, make, etc.
RUN apk add \
    --update \
    --no-cache \
    alpine-sdk vips vips-dev

ENV GO111MODULE on

WORKDIR /go/src/github.com/skygeario/skygear-server

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN cd cmd/asset; make build; mv skygear-asset /tmp/

FROM alpine:3.11

# Install mailcap to get /etc/mime.types
# Golang uses that file to augment mime package.
# See https://golang.org/pkg/mime/#TypeByExtension
RUN apk add \
    --update \
    --no-cache \
    ca-certificates mailcap vips

COPY --from=godev /tmp/skygear-asset /
RUN chmod a+x /skygear-asset
USER nobody

EXPOSE 3002

CMD ["/skygear-asset"]
