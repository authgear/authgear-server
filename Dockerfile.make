FROM golang:1.9.4-stretch as godev

RUN \
    apt-get update && \
    apt-get install --no-install-recommends -y libtool-bin automake pkg-config libsodium-dev libzmq3-dev && \
    rm -rf /var/lib/apt/lists/* && \
    go get github.com/golang/dep/cmd/dep

RUN mkdir -p /go/src/github.com/skygeario/skygear-server
WORKDIR /go/src/github.com/skygeario/skygear-server
SHELL ["/bin/bash", "-c"]

# Copy a minimal set of files to restore Go dependencies to get advantage
# of Docker build cache
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only

RUN \
    cp -rf vendor/* /go/src; \
    cd /go/src; \
    for pkg in \
        "golang.org/x/tools/cmd/stringer" \
        "golang.org/x/tools/cmd/cover" \
        "github.com/golang/lint/golint" \
        "github.com/rickmak/gocyclo" \
        "github.com/oursky/gogocyclo" \
        "github.com/mitchellh/gox" \
        "github.com/golang/mock/mockgen" \
        "honnef.co/go/tools/cmd/staticcheck" \
        ; do \
        pushd $pkg; \
        go install .; \
        popd; \
    done; \
