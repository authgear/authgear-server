FROM golang:1.10.4-stretch as godev

WORKDIR /go/src/github.com/skygeario/skygear-server
RUN go get github.com/golang/dep/cmd/dep

COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only

COPY . .
RUN cd cmd/auth; make build; mv skygear-auth /tmp/

FROM alpine:3.8

RUN apk add --update --no-cache libc6-compat ca-certificates && \
    ln -s /lib /lib64

COPY --from=godev /tmp/skygear-auth /
RUN chmod a+x /skygear-auth
USER nobody

EXPOSE 3000

CMD ["/skygear-auth"]
