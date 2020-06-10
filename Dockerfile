FROM golang:1.13.3-stretch as godev

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build

FROM alpine:3.8

RUN apk add --update --no-cache libc6-compat ca-certificates && \
    ln -s /lib /lib64

COPY ./reserved_name.txt /reserved_name.txt

COPY --from=godev /src/authgear /
USER nobody

EXPOSE 3000

CMD ["/authgear"]
