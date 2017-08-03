FROM golang:1.8.3-stretch

RUN \
    apt-get update && \
    apt-get install --no-install-recommends -y libtool-bin automake pkg-config libsodium-dev libzmq3-dev libczmq-dev && \
    rm -rf /var/lib/apt/lists/*

RUN curl https://glide.sh/get | sh && \
    rm -rf /tmp/glide*

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

RUN go get golang.org/x/tools/cmd/stringer \
           golang.org/x/tools/cmd/cover \
           github.com/golang/lint/golint \
           github.com/golang/mock/mockgen \
           github.com/rickmak/gocyclo \
           github.com/oursky/gogocyclo \
           github.com/mitchellh/gox \
           github.com/vektra/mockery/.../ \
           github.com/derekparker/delve/cmd/dlv \
           honnef.co/go/tools/cmd/staticcheck
