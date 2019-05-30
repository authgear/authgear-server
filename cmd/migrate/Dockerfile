FROM golang:1.12.5-stretch as godev

ENV GO111MODULE on

WORKDIR /go/src/github.com/skygeario/skygear-server

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN cd cmd/migrate; make build; mv skygear-migrate /tmp/

FROM alpine:3.8

RUN apk add --update --no-cache libc6-compat ca-certificates && \
    ln -s /lib /lib64

RUN mkdir -p cmd/migrate/revisions
COPY cmd/migrate/revisions cmd/migrate/revisions

COPY --from=godev /tmp/skygear-migrate /
RUN chmod a+x /skygear-migrate
USER nobody

ENTRYPOINT ["/skygear-migrate"]
