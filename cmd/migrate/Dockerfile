FROM golang:1.10.4-stretch as godev

WORKDIR /go/src/github.com/skygeario/skygear-server
RUN go get github.com/golang/dep/cmd/dep

COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only

COPY . .
RUN cd cmd/migrate; make build; mv skygear-migrate /tmp/

FROM alpine:3.4

RUN apk add --update --no-cache libc6-compat ca-certificates && \
    ln -s /lib /lib64

RUN mkdir -p cmd/migrate/revisions
COPY cmd/migrate/revisions cmd/migrate/revisions

COPY --from=godev /tmp/skygear-migrate /
RUN chmod a+x /skygear-migrate
USER nobody

ENTRYPOINT ["/skygear-migrate"]
