FROM golang:onbuild

RUN \
    go get golang.org/x/tools/cmd/cover && \
    go get github.com/golang/lint/golint && \
    go get github.com/smartystreets/goconvey/convey && \
    go get github.com/smartystreets/assertions && \
    go get golang.org/x/tools/cmd/stringer && \
    golint ./... && \
    go generate ./...

EXPOSE 3000


