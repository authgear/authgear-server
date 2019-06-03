FROM golang:1.12.5-stretch

ENV GO111MODULE on
SHELL ["/bin/bash", "-c"]

WORKDIR /go/src/github.com/skygeario/skygear-server

RUN curl -fsSL https://github.com/alecthomas/gometalinter/releases/download/v2.0.11/gometalinter-2.0.11-linux-amd64.tar.gz | tar --strip-components 1 -C /usr/local/bin -zx

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
