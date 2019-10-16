FROM golang:1.13.3-stretch as godev

ENV GO111MODULE on

WORKDIR /go/src/github.com/skygeario/skygear-server

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN cd cmd/gateway; make build; mv skygear-gateway /tmp/

FROM alpine:3.8

RUN apk add --update --no-cache libc6-compat ca-certificates && \
    ln -s /lib /lib64

COPY --from=godev /tmp/skygear-gateway /
RUN chmod a+x /skygear-gateway
USER nobody

EXPOSE 3001

CMD ["/skygear-gateway"]
