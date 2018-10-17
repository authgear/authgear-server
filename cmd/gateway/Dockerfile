FROM golang:1.10.4-stretch as godev

WORKDIR /go/src/github.com/skygeario/skygear-server
RUN go get github.com/golang/dep/cmd/dep

COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only

COPY . .
RUN cd cmd/gateway; make build; mv skygear-gateway /tmp/

FROM alpine:3.4

RUN apk add --update --no-cache libc6-compat ca-certificates && \
    ln -s /lib /lib64

COPY --from=godev /tmp/skygear-gateway /
RUN chmod a+x /skygear-gateway
USER nobody

EXPOSE 3001

CMD ["/skygear-gateway"]
