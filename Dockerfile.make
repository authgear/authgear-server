FROM golang:1.12.5-stretch

ENV GO111MODULE on

WORKDIR /go/src/github.com/skygeario/skygear-server
SHELL ["/bin/bash", "-c"]

COPY go.mod go.sum ./
RUN go mod download

ENV GOBIN /go/bin
COPY ./tools/nextimportslint.go ./tools/nextimportslint.go

RUN \
    cp -rf vendor/* /go/src; \
    cd /go/src; \
    for pkg in \
        "golang.org/x/tools/cmd/stringer" \
        "golang.org/x/tools/cmd/cover" \
        "github.com/tinylib/msgp" \
        "github.com/mitchellh/gox" \
        "github.com/golang/mock/mockgen" \
        ; do \
        pushd $pkg; \
        go install .; \
        popd > /dev/null; \
    done; \
    go install ./github.com/skygeario/skygear-server/tools/nextimportslint.go; \
    curl -fsSL https://github.com/alecthomas/gometalinter/releases/download/v2.0.11/gometalinter-2.0.11-linux-amd64.tar.gz | tar --strip-components 1 -C /usr/local/bin -zx;
