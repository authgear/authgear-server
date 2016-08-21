DIST_DIR = ./dist/
DIST := skygear-server
VERSION := $(shell git describe --always --tags)
GO_BUILD_LDFLAGS := -ldflags "-X github.com/skygeario/skygear-server/pkg/server/skyversion.version=$(VERSION)"
OSARCHS := linux/amd64 linux/386 linux/arm windows/amd64 windows/386 darwin/amd64

ifeq (1,${WITH_ZMQ})
GO_BUILD_TAGS := --tags zmq
endif

ifeq (1,${WITH_DOCKER})
DOCKER_RUN := docker run --rm -i \
	-v `pwd`:/go/src/github.com/skygeario/skygear-server \
	-w /go/src/github.com/skygeario/skygear-server \
	skygeario/skygear-godev:latest
endif

GO_BUILD_ARGS := $(GO_BUILD_TAGS) $(GO_BUILD_LDFLAGS)

.PHONY: vendor
vendor:
	$(DOCKER_RUN) glide install

.PHONY: build
build:
	$(DOCKER_RUN) go build -o $(DIST) $(GO_BUILD_ARGS)

.PHONY: clean
	rm -rf $(DIST_DIR)

.PHONY: all
all:
	mkdir -p $(DIST_DIR)
	$(DOCKER_RUN) gox -osarch="$(OSARCHS)" -output="$(DIST_DIR)/{{.Dir}}-{{.OS}}-{{.Arch}}" $(GO_BUILD_ARGS)
	$(MAKE) build GOOS=linux GOARCH=amd64 DIST=$(DIST_DIR)$(DIST)-zmq-linux-amd64 WITH_ZMQ=1

.PHONY: docker-build
docker-build: build
	cp skygear-server scripts/release/
	make -C scripts/release docker-build

.PHONY: docker-build
docker-push:
	make -C scripts/release docker-push
