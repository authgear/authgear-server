FROM golang:1.13.3-stretch as godev

ENV GO111MODULE on

WORKDIR /go/src/github.com/skygeario/skygear-server

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN cd cmd/auth; make build; mv skygear-auth /tmp/

FROM alpine:3.8

RUN apk add --update --no-cache libc6-compat ca-certificates && \
    ln -s /lib /lib64

COPY ./reserved_name.txt /reserved_name.txt

COPY --from=godev /tmp/skygear-auth /
RUN chmod a+x /skygear-auth
USER nobody

EXPOSE 3000

CMD ["/skygear-auth"]
