DIST_DIR = ./dist/
DIST := skygear-server
VERSION := $(shell git describe --always)
GIT_SHA := $(shell git rev-parse HEAD)
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
TARGETS := gateway auth
GO_BUILD_LDFLAGS := -ldflags "-X github.com/skygeario/skygear-server/pkg/server/skyversion.version=$(VERSION)"
GO_TEST_TIMEOUT := 1m30s
OSARCHS := linux/amd64 linux/386 linux/arm windows/amd64 windows/386 darwin/amd64
GO_TEST_CPU := 1,4
GO_TEST_PACKAGE := ./pkg/core/... ./pkg/auth/... ./pkg/gateway/...
SHELL := /bin/bash

ifeq (1,${GO_TEST_VERBOSE})
GO_TEST_ARGS_VERBOSE := -v
endif

DOCKER_REGISTRY :=
DOCKER_ORG_NAME := skygeario
DOCKER_IMAGE_AUTH := skygear-auth
DOCKER_IMAGE_GATEWAY := skygear-gateway
DOCKER_TAG := git-$(shell git rev-parse --short HEAD)
PUSH_DOCKER_TAG := $(VERSION)
IMAGE_NAME = $(DOCKER_REGISTRY)$(DOCKER_ORG_NAME)/$(1):$(DOCKER_TAG)
VERSIONED_IMAGE_NAME = $($(DOCKER_REGISTRY)$(DOCKER_ORG_NAME)/$(1):$(PUSH_DOCKER_TAG))
AUTH_IMAGE_NAME = $(call IMAGE_NAME,$(DOCKER_IMAGE_AUTH))
AUTH_VERSIONED_IMAGE_NAME = $(call VERSIONED_IMAGE_NAME,$(DOCKER_IMAGE_AUTH))
GATEWAY_IMAGE_NAME = $(call IMAGE_NAME,$(DOCKER_IMAGE_GATEWAY))
GATEWAY_VERSIONED_IMAGE_NAME = $(call VERSIONED_IMAGE_NAME,$(DOCKER_IMAGE_GATEWAY))

GO_BUILD_ARGS := $(GO_BUILD_TAGS) $(GO_BUILD_LDFLAGS)
GO_TEST_ARGS := $(GO_BUILD_ARGS) -cover -timeout $(GO_TEST_TIMEOUT) $(GO_TEST_ARGS_VERBOSE) -p 1 -cpu $(GO_TEST_CPU)

.PHONY: vendor
vendor:
	dep ensure

.PHONY: go-install
go-install:
	go install $(GO_BUILD_ARGS) ./...
	go install tools/nextimportslint.go

.PHONY: go-generate
go-generate: go-install
	find pkg -type f -name "*_gen.go" -delete
	find pkg -type f -name "mockgen_*.go" -delete
	go generate ./pkg/...

.PHONY: go-lint
go-lint: go-install
	gometalinter --disable-all \
		-enable=staticcheck --enable=golint --enable=misspell --enable=gocyclo \
		--linter='gocyclo:gocyclo -over 15:^(?P<cyclo>\d+)\s+\S+\s(?P<function>\S+)\s+(?P<path>.*?\.go):(?P<line>\d+):(\d+)$'' \
		./...
# Next linter have stricter rule
	gometalinter ./pkg/auth/... ./pkg/core/... ./pkg/gateway/...
	nextimportslint

.PHONY: generate
generate: go-generate

.PHONY: build
build:
	go build -o $(DIST) $(GO_BUILD_ARGS)
	chmod +x $(DIST)

.PHONY: test
test:
# Run `go install` to compile packages for caching and catch compilation error.
	for TARGET in $(TARGETS) ; do \
		pushd cmd/$$TARGET > /dev/null ; \
		go install $(GO_BUILD_ARGS) ; \
		popd > /dev/null ; \
	done
	go test $(GO_TEST_ARGS) $(GO_TEST_PACKAGE)

.PHONY: lint
lint: go-lint

.PHONY: fmt
fmt:
	gofmt -w main.go ./pkg

.PHONY: clean
clean:
	rm -rf $(DIST_DIR)

.PHONY: all
all:
	for TARGET in $(TARGETS) ; do \
		DIST_DIR=$(DIST_DIR)$$TARGET/ ; \
		mkdir -p $$DIST_DIR ; \
		cp cmd/$$TARGET/main.go . ; \
		gox -osarch="$(OSARCHS)" -output="$$DIST_DIR/$$TARGET-{{.OS}}-{{.Arch}}" $(GO_BUILD_ARGS) ; \
		$(MAKE) build GOOS=linux GOARCH=amd64 DIST=$$DIST_DIR$$TARGET-linux-amd64; \
		chmod +x $$DIST_DIR$$TARGET* ; \
		rm main.go ; \
	done

.PHONY: update-version
update-version:
	sed -i "" "s/version = \".*\"/version = \"v$(SKYGEAR_VERSION)\"/" pkg/server/skyversion/version.go

.PHONY: archive
archive:
	cd $(DIST_DIR) ; \
		find . -maxdepth 2 -type f \( -name 'auth-*' -o -name 'gateway-*' \) -not -name '*.exe' -not -name '*.zip' -not -name '*.tar.gz' -exec tar -zcvf {}.tar.gz {} \; ; \
		find . -maxdepth 2 -type f \( -name 'auth-*.exe' -o -name 'gateway-*.exe' \) -not -exec zip -r {}.zip {} \;

.PHONY: docker-build-image
docker-build-image:
	docker build \
	  -f $(DOCKER_FILE) \
		-t $(IMAGE_NAME) \
		--build-arg sha=$(GIT_SHA) \
		--build-arg version=$(VERSION) \
		--build-arg build_date=$(BUILD_DATE) \
		.

.PHONY: docker-build-auth
docker-build-auth:
	$(MAKE) docker-build-image DOCKER_FILE=cmd/auth/Dockerfile IMAGE_NAME=$(AUTH_IMAGE_NAME)

.PHONY: docker-build-gateway
docker-build-gateway:
	$(MAKE) docker-build-image DOCKER_FILE=cmd/auth/Dockerfile IMAGE_NAME=$(GATEWAY_IMAGE_NAME)

.PHONY: docker-build
docker-build: docker-build-auth docker-build-gateway

.PHONY: docker-push
docker-push:
	docker push $(AUTH_IMAGE_NAME)
	docker push $(GATEWAY_IMAGE_NAME)

.PHONY: docker-push-version
docker-push-version:
	docker tag $(AUTH_IMAGE_NAME) $(AUTH_VERSIONED_IMAGE_NAME)
	docker tag $(GATEWAY_IMAGE_NAME) $(GATEWAY_VERSIONED_IMAGE_NAME)
	docker push $(AUTH_VERSIONED_IMAGE_NAME)
	docker push $(GATEWAY_VERSIONED_IMAGE_NAME)

.PHONY: release-commit
release-commit:
	./scripts/release-commit.sh
