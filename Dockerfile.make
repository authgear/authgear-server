FROM golang:1.13.3-stretch

ENV GO111MODULE on
SHELL ["/bin/bash", "-c"]

WORKDIR /go/src/github.com/skygeario/skygear-server

RUN curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b $(go env GOPATH)/bin v1.22.2

COPY go.mod go.sum ./
RUN go mod download

COPY ./tools/nextimportslint.go ./tools/nextimportslint.go

RUN go mod download
RUN go install golang.org/x/tools/cmd/stringer
RUN go install golang.org/x/tools/cmd/cover
RUN go install github.com/tinylib/msgp
RUN go install github.com/mitchellh/gox
RUN go install github.com/golang/mock/mockgen
RUN go install tools/nextimportslint.go
