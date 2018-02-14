FROM golang:1.9.4-stretch as godev

RUN \
    apt-get update && \
    apt-get install --no-install-recommends -y libtool-bin automake pkg-config libsodium-dev libzmq3-dev && \
    rm -rf /var/lib/apt/lists/* && \
    curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 && \
    curl -fsSL -o /usr/local/bin/vg https://github.com/GetStream/vg/releases/download/v0.8.0/vg-linux-amd64 && \
    chmod +x /usr/local/bin/dep /usr/local/bin/vg && \
    curl -fsSL https://github.com/alecthomas/gometalinter/releases/download/v2.0.4/gometalinter-2.0.4-linux-amd64.tar.gz | tar --strip-components 1 -C /usr/local/bin -zx

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
        "github.com/mitchellh/gox" \
        "github.com/golang/mock/mockgen" \
        ; do \
        pushd $pkg; \
        go install .; \
        popd; \
    done; \
