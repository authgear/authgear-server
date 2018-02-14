DIST_DIR = ./dist/
DIST := skygear-server
VERSION := $(shell git describe --always)
GIT_SHA := $(shell git rev-parse HEAD)
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_BUILD_LDFLAGS := -ldflags "-X github.com/skygeario/skygear-server/pkg/server/skyversion.version=$(VERSION)"
GO_TEST_TIMEOUT := 1m30s
OSARCHS := linux/amd64 linux/386 linux/arm windows/amd64 windows/386 darwin/amd64
GO_TEST_CPU := 1,4
GO_TEST_PACKAGE := ./pkg/...
SHELL := /bin/bash

ifeq (1,${WITH_ZMQ})
GO_BUILD_TAGS := --tags zmq
endif

DOCKER_COMPOSE_CMD := docker-compose \
	-f docker-compose.make.yml \

DOCKER_COMPOSE_CMD_TEST := docker-compose \
	-f docker-compose.test.yml \
	-p skygear-server-test

ifeq (1,${WITH_DOCKER})
DOCKER_RUN := ${DOCKER_COMPOSE_CMD} run --rm app
DOCKER_RUN_DB := ${DOCKER_COMPOSE_CMD_TEST} run --rm db_cmd
DOCKER_RUN_TEST := ${DOCKER_COMPOSE_CMD_TEST} run --rm app
GO_TEST_TIMEOUT := 5m
endif

DOCKER_REGISTRY :=
DOCKER_ORG_NAME := skygeario
DOCKER_IMAGE := skygear-server
DOCKER_TAG := git-$(shell git rev-parse --short HEAD)
PUSH_DOCKER_TAG := $(VERSION)
IMAGE_NAME := $(DOCKER_REGISTRY)$(DOCKER_ORG_NAME)/$(DOCKER_IMAGE):$(DOCKER_TAG)

GO_BUILD_ARGS := $(GO_BUILD_TAGS) $(GO_BUILD_LDFLAGS)

.PHONY: vendor
vendor:
	$(DOCKER_RUN) dep ensure

.PHONY: go-install
go-install:
	$(DOCKER_RUN) go install $(GO_BUILD_ARGS) ./...

.PHONY: go-generate
go-generate: go-install
	$(DOCKER_RUN) find pkg -type f -name "mock_*.go" -delete
	$(DOCKER_RUN) go generate ./pkg/...

.PHONY: go-lint
go-lint: go-install
	$(DOCKER_RUN) gometalinter --disable-all --enable=gocyclo --enable=staticcheck --enable=golint --enable=misspell ./...
	$(DOCKER_RUN) gometalinter ./... || true

.PHONY: generate
generate: go-generate

.PHONY: build
build:
	$(DOCKER_RUN) go build -o $(DIST) $(GO_BUILD_ARGS)
	$(DOCKER_RUN) chmod +x $(DIST)

.PHONY: before-docker-test
before-docker-test:
	-$(DOCKER_COMPOSE_CMD_TEST) up -d db redis
	sleep 20
	make before-test WITH_DOCKER=1

.PHONY: before-test
before-test:
	-$(DOCKER_RUN_DB) psql -c 'CREATE DATABASE skygear_test;'

.PHONY: test
test:
# Run `go install` to compile packages for caching and catch compilation error.
	$(DOCKER_RUN_TEST) go install $(GO_BUILD_ARGS)
	$(DOCKER_RUN_TEST) go test $(GO_BUILD_ARGS) -cover -timeout $(GO_TEST_TIMEOUT) -p 1 -cpu $(GO_TEST_CPU) $(GO_TEST_PACKAGE)

.PHONY: lint
lint: go-lint

.PHONY: fmt
fmt:
	$(DOCKER_RUN) gofmt -w main.go ./pkg

.PHONY: after-docker-test
after-docker-test:
	-$(DOCKER_COMPOSE_CMD_TEST) down -v

.PHONY: clean
clean:
	rm -rf $(DIST_DIR)

.PHONY: all
all:
	mkdir -p $(DIST_DIR)
	$(DOCKER_RUN) gox -osarch="$(OSARCHS)" -output="$(DIST_DIR)/{{.Dir}}-{{.OS}}-{{.Arch}}" $(GO_BUILD_ARGS)
	$(MAKE) build GOOS=linux GOARCH=amd64 DIST=$(DIST_DIR)$(DIST)-zmq-linux-amd64 WITH_ZMQ=1
	$(DOCKER_RUN) chmod +x $(DIST_DIR)/$(DIST)*

.PHONY: update-version
update-version:
	sed -i "" "s/version = \".*\"/version = \"v$(SKYGEAR_VERSION)\"/" pkg/server/skyversion/version.go

.PHONY: archive
archive:
	cd $(DIST_DIR) ; \
		find . -maxdepth 1 -type f -name 'skygear-server-*' -not -name '*.exe' -not -name '*.zip' -not -name '*.tar.gz' -exec tar -zcvf {}.tar.gz {} \; ; \
		find . -maxdepth 1 -type f -name 'skygear-server-*.exe' -not -exec zip -r {}.zip {} \;

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE_NAME) \
		--build-arg sha=$(GIT_SHA) \
		--build-arg version=$(VERSION) \
		--build-arg build_date=$(BUILD_DATE) \
		.

.PHONY: docker-push
docker-push:
	docker push $(IMAGE_NAME)

.PHONY: docker-push-version
docker-push-version:
	docker tag $(IMAGE_NAME) $(DOCKER_REGISTRY)$(DOCKER_ORG_NAME)/$(DOCKER_IMAGE):$(PUSH_DOCKER_TAG)
	docker push $(DOCKER_REGISTRY)$(DOCKER_ORG_NAME)/$(DOCKER_IMAGE):$(PUSH_DOCKER_TAG)

.PHONY: release-commit
release-commit:
	./scripts/release-commit.sh
