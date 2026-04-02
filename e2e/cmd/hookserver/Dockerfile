FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY main.go .
RUN go build -o /out/e2e-hook .

FROM alpine:3.22

COPY --from=build /out/e2e-hook /usr/local/bin/e2e-hook

EXPOSE 2626

ENTRYPOINT ["/usr/local/bin/e2e-hook"]
