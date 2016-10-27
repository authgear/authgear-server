DIST_DIR = ./dist/
DIST := skygear-server
VERSION := $(shell git describe --always --tags)
GO_BUILD_LDFLAGS := -ldflags "-X github.com/skygeario/skygear-server/pkg/server/skyversion.version=$(VERSION)"
GO_TEST_TIMEOUT := 1m
OSARCHS := linux/amd64 linux/386 linux/arm windows/amd64 windows/386 darwin/amd64

ifeq (1,${WITH_ZMQ})
GO_BUILD_TAGS := --tags zmq
endif

DOCKER_COMPOSE_CMD := docker-compose \
	-f docker-compose.dev.yml \
	-p skygear-server-test

ifeq (1,${WITH_DOCKER})
DOCKER_RUN := docker run --rm -i \
	-v `pwd`:/go/src/github.com/skygeario/skygear-server \
	-w /go/src/github.com/skygeario/skygear-server \
	skygeario/skygear-godev:latest
DOCKER_COMPOSE_RUN := ${DOCKER_COMPOSE_CMD} run --rm app
DOCKER_COMPOSE_RUN_DB := ${DOCKER_COMPOSE_CMD} run --rm db_cmd
GO_TEST_TIMEOUT := 5m
endif

GO_BUILD_ARGS := $(GO_BUILD_TAGS) $(GO_BUILD_LDFLAGS)

.PHONY: vendor
vendor:
	$(DOCKER_RUN) glide install

.PHONY: build
build:
	$(DOCKER_RUN) go build -o $(DIST) $(GO_BUILD_ARGS)
	$(DOCKER_RUN) chmod +x $(DIST)

.PHONY: before-docker-test
before-docker-test:
	-$(DOCKER_COMPOSE_CMD) up -d db redis
	sleep 20
	make before-test WITH_DOCKER=1

.PHONY: before-test
before-test:
	-$(DOCKER_COMPOSE_RUN_DB) psql -c 'CREATE DATABASE skygear_test;'

.PHONY: test
test:
# Run `go install` to compile packages for caching and catch compilation error.
	$(DOCKER_COMPOSE_RUN) go install $(GO_BUILD_ARGS)
# The pq test suites do not run well without other test suites, so they are run
# separately.
	$(DOCKER_COMPOSE_RUN) go test $(GO_BUILD_ARGS) -cover -timeout $(GO_TEST_TIMEOUT) -cpu 1 ./pkg/server/skydb/pq/...
# Run the test of test suites. pq test suites are skipped when GOMAXPROCS != 1.
	$(DOCKER_COMPOSE_RUN) go test $(GO_BUILD_ARGS) -cover -timeout $(GO_TEST_TIMEOUT) -cpu 4 ./pkg/...

.PHONY: after-docker-test
after-docker-test:
	-$(DOCKER_COMPOSE_CMD) down -v

.PHONY: clean
	rm -rf $(DIST_DIR)

.PHONY: all
all:
	mkdir -p $(DIST_DIR)
	$(DOCKER_RUN) gox -osarch="$(OSARCHS)" -output="$(DIST_DIR)/{{.Dir}}-{{.OS}}-{{.Arch}}" $(GO_BUILD_ARGS)
	$(MAKE) build GOOS=linux GOARCH=amd64 DIST=$(DIST_DIR)$(DIST)-zmq-linux-amd64 WITH_ZMQ=1
	$(DOCKER_RUN) chmod +x $(DIST_DIR)/$(DIST)*

.PHONY: docker-build
docker-build: build
	cp skygear-server scripts/release/
	make -C scripts/release docker-build

.PHONY: docker-build
docker-push:
	make -C scripts/release docker-push
