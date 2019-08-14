VERSION := $(shell git describe --always)
GIT_SHA := $(shell git rev-parse HEAD)
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_BUILD_LDFLAGS := -ldflags "-X github.com/skygeario/skygear-server/pkg/server/skyversion.version=$(VERSION)"
GO_TEST_TIMEOUT := 1m30s
GO_TEST_CPU := 1,4
GO_TEST_PACKAGE := ./pkg/core/... ./pkg/auth/... ./pkg/gateway/...
SHELL := /bin/bash

ifeq (1,${GO_TEST_VERBOSE})
GO_TEST_ARGS_VERBOSE := -v
endif

DOCKER_COMPOSE_CMD := docker-compose -f docker-compose.make.yml

ifeq (1,${WITH_DOCKER})	
DOCKER_RUN := ${DOCKER_COMPOSE_CMD} run --rm app
endif

DOCKER_REGISTRY :=
DOCKER_ORG_NAME := skygeario
DOCKER_IMAGE_AUTH := skygear-auth
DOCKER_IMAGE_GATEWAY := skygear-gateway
DOCKER_IMAGE_MIGRATE := skygear-migrate
DOCKER_TAG := git-$(shell git rev-parse --short HEAD)
PUSH_DOCKER_TAG := $(VERSION)
IMAGE_NAME = $(DOCKER_REGISTRY)$(DOCKER_ORG_NAME)/$(1):$(DOCKER_TAG)
VERSIONED_IMAGE_NAME = $(DOCKER_REGISTRY)$(DOCKER_ORG_NAME)/$(1):$(PUSH_DOCKER_TAG)
AUTH_IMAGE_NAME = $(call IMAGE_NAME,$(DOCKER_IMAGE_AUTH))
AUTH_VERSIONED_IMAGE_NAME = $(call VERSIONED_IMAGE_NAME,$(DOCKER_IMAGE_AUTH))
GATEWAY_IMAGE_NAME = $(call IMAGE_NAME,$(DOCKER_IMAGE_GATEWAY))
GATEWAY_VERSIONED_IMAGE_NAME = $(call VERSIONED_IMAGE_NAME,$(DOCKER_IMAGE_GATEWAY))
MIGRATE_IMAGE_NAME = $(call IMAGE_NAME,$(DOCKER_IMAGE_MIGRATE))
MIGRATE_VERSIONED_IMAGE_NAME = $(call VERSIONED_IMAGE_NAME,$(DOCKER_IMAGE_MIGRATE))

GO_BUILD_ARGS := $(GO_BUILD_TAGS) $(GO_BUILD_LDFLAGS)
GO_TEST_ARGS := $(GO_BUILD_ARGS) -cover -timeout $(GO_TEST_TIMEOUT) $(GO_TEST_ARGS_VERBOSE) -p 1 -cpu $(GO_TEST_CPU)

.PHONY: vendor
vendor:
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b $$(go env GOPATH)/bin v1.16.0
	go mod download
	go install golang.org/x/tools/cmd/stringer
	go install golang.org/x/tools/cmd/cover
	go install github.com/tinylib/msgp
	go install github.com/mitchellh/gox
	go install github.com/golang/mock/mockgen
	go install tools/nextimportslint.go

.PHONY: go-generate
go-generate:
	$(DOCKER_RUN) find pkg -type f -name "*_gen.go" -delete
	$(DOCKER_RUN) find pkg -type f -name "mockgen_*.go" -delete
	$(DOCKER_RUN) go generate ./pkg/...

.PHONY: go-lint
go-lint:
	$(DOCKER_RUN) golangci-lint run ./cmd/... ./pkg/...
	$(DOCKER_RUN) nextimportslint

.PHONY: generate
generate: go-generate

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: test
test:
# Run `go install` to compile packages for caching and catch compilation error.
	$(DOCKER_RUN) go install $(GO_BUILD_ARGS) ./cmd/...
	$(DOCKER_RUN) go test $(GO_TEST_ARGS) $(GO_TEST_PACKAGE)

.PHONY: lint
lint: go-lint

.PHONY: fmt
fmt:
	${DOCKER_RUN} gofmt -w cmd/**/main.go ./pkg

.PHONY: update-version
update-version:
	sed -i "" "s/version = \".*\"/version = \"v$(SKYGEAR_VERSION)\"/" pkg/server/skyversion/version.go

.PHONY: docker-build-image
docker-build-image: CONTEXT ?= .
docker-build-image:
	docker build \
	  -f $(DOCKER_FILE) \
		-t $(IMAGE_NAME) \
		--build-arg sha=$(GIT_SHA) \
		--build-arg version=$(VERSION) \
		--build-arg build_date=$(BUILD_DATE) \
		$(CONTEXT)

.PHONY: docker-build-auth
docker-build-auth:
	$(MAKE) docker-build-image DOCKER_FILE=cmd/auth/Dockerfile IMAGE_NAME=$(AUTH_IMAGE_NAME)

.PHONY: docker-build-gateway
docker-build-gateway:
	$(MAKE) docker-build-image DOCKER_FILE=cmd/gateway/Dockerfile IMAGE_NAME=$(GATEWAY_IMAGE_NAME)

.PHONY: docker-build-migrate
docker-build-migrate:
	$(MAKE) docker-build-image DOCKER_FILE=./migrate/cmd/migrate/Dockerfile IMAGE_NAME=$(MIGRATE_IMAGE_NAME) CONTEXT=./migrate

.PHONY: docker-build
docker-build: docker-build-auth docker-build-gateway docker-build-migrate

.PHONY: docker-push
docker-push:
	docker push $(AUTH_IMAGE_NAME)
	docker push $(GATEWAY_IMAGE_NAME)
	docker push $(MIGRATE_IMAGE_NAME)

.PHONY: docker-push-version
docker-push-version:
	docker tag $(AUTH_IMAGE_NAME) $(AUTH_VERSIONED_IMAGE_NAME)
	docker tag $(GATEWAY_IMAGE_NAME) $(GATEWAY_VERSIONED_IMAGE_NAME)
	docker tag $(MIGRATE_IMAGE_NAME) $(MIGRATE_VERSIONED_IMAGE_NAME)
	docker push $(AUTH_VERSIONED_IMAGE_NAME)
	docker push $(GATEWAY_VERSIONED_IMAGE_NAME)
	docker push $(MIGRATE_VERSIONED_IMAGE_NAME)

.PHONY: release-commit
release-commit:
	./scripts/release-commit.sh

.PHONY: preview-doc-auth
preview-doc-auth:
	./scripts/preview-doc.sh auth

.PHONY: generate-doc-auth
generate-doc-auth:
	@openapi3-gen -output "$(DOC_PATH)" ./cmd/auth/... ./pkg/auth/...

# Start the migrate http server locally
.PHONY: serve-migrate-http
serve-migrate-http:
	go run ./cmd/migrate/main.go -add-migration-src=auth,cmd/migrate/revisions/auth -add-migration-src=core,cmd/migrate/revisions/core -http-server
