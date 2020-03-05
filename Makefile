VERSION := $(shell git describe --always)
GIT_SHA := $(shell git rev-parse HEAD)
GIT_SHORT_SHA := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_CONTEXT ?= .
GO_BUILD_LDFLAGS := -ldflags "-X github.com/skygeario/skygear-server/pkg/server/skyversion.version=$(VERSION)"
GO_TEST_TIMEOUT := 1m30s
GO_TEST_PACKAGE := ./pkg/core/... ./pkg/auth/... ./pkg/gateway/... ./pkg/asset/...
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
DOCKER_IMAGE_ASSET := skygear-asset
DOCKER_IMAGE_MIGRATE := skygear-migrate
DOCKER_TAG := git-$(shell git rev-parse --short HEAD)
PUSH_DOCKER_TAG := $(VERSION)
IMAGE_NAME = $(DOCKER_REGISTRY)$(DOCKER_ORG_NAME)/$(1):$(DOCKER_TAG)
VERSIONED_IMAGE_NAME = $(DOCKER_REGISTRY)$(DOCKER_ORG_NAME)/$(1):$(PUSH_DOCKER_TAG)
AUTH_IMAGE_NAME = $(call IMAGE_NAME,$(DOCKER_IMAGE_AUTH))
AUTH_VERSIONED_IMAGE_NAME = $(call VERSIONED_IMAGE_NAME,$(DOCKER_IMAGE_AUTH))
GATEWAY_IMAGE_NAME = $(call IMAGE_NAME,$(DOCKER_IMAGE_GATEWAY))
GATEWAY_VERSIONED_IMAGE_NAME = $(call VERSIONED_IMAGE_NAME,$(DOCKER_IMAGE_GATEWAY))
ASSET_IMAGE_NAME = $(call IMAGE_NAME,$(DOCKER_IMAGE_ASSET))
ASSET_VERSIONED_IMAGE_NAME = $(call VERSIONED_IMAGE_NAME,$(DOCKER_IMAGE_ASSET))
MIGRATE_IMAGE_NAME = $(call IMAGE_NAME,$(DOCKER_IMAGE_MIGRATE))
MIGRATE_VERSIONED_IMAGE_NAME = $(call VERSIONED_IMAGE_NAME,$(DOCKER_IMAGE_MIGRATE))

GO_BUILD_ARGS := $(GO_BUILD_TAGS) $(GO_BUILD_LDFLAGS)
GO_TEST_ARGS := $(GO_BUILD_ARGS) -cover -timeout $(GO_TEST_TIMEOUT) $(GO_TEST_ARGS_VERBOSE)

.PHONY: vendor
vendor:
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b $$(go env GOPATH)/bin v1.22.2
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
	sed -i "" "s/const MajorVersion = .*/const MajorVersion = $(shell echo $$SKYGEAR_VERSION | cut -d . -f 1)/" pkg/core/apiversion/apiversion.go
	sed -i "" "s/const MinorVersion = .*/const MinorVersion = $(shell echo $$SKYGEAR_VERSION | cut -d . -f 2)/" pkg/core/apiversion/apiversion.go

.PHONY: docker-build-image
docker-build-image:
	docker build \
	  -f $(DOCKER_FILE) \
		-t $(IMAGE_NAME) \
		--build-arg sha=$(GIT_SHA) \
		--build-arg version=$(VERSION) \
		--build-arg build_date=$(BUILD_DATE) \
		$(BUILD_CONTEXT)

.PHONY: docker-build-auth
docker-build-auth:
	$(MAKE) docker-build-image DOCKER_FILE=cmd/auth/Dockerfile IMAGE_NAME=$(AUTH_IMAGE_NAME)

.PHONY: docker-build-gateway
docker-build-gateway:
	$(MAKE) docker-build-image DOCKER_FILE=cmd/gateway/Dockerfile IMAGE_NAME=$(GATEWAY_IMAGE_NAME)

.PHONY: docker-build-asset
docker-build-asset:
	$(MAKE) docker-build-image DOCKER_FILE=cmd/asset/Dockerfile IMAGE_NAME=$(ASSET_IMAGE_NAME)

.PHONY: docker-build-migrate
docker-build-migrate:
	$(MAKE) docker-build-image DOCKER_FILE=./migrate/cmd/migrate/Dockerfile IMAGE_NAME=$(MIGRATE_IMAGE_NAME) BUILD_CONTEXT=./migrate

.PHONY: docker-build
docker-build: docker-build-auth docker-build-gateway docker-build-migrate docker-build-asset

.PHONY: docker-push
docker-push:
	docker push $(AUTH_IMAGE_NAME)
	docker push $(GATEWAY_IMAGE_NAME)
	docker push $(ASSET_IMAGE_NAME)
	docker push $(MIGRATE_IMAGE_NAME)

.PHONY: docker-push-version
docker-push-version:
	docker tag $(AUTH_IMAGE_NAME) $(AUTH_VERSIONED_IMAGE_NAME)
	docker tag $(GATEWAY_IMAGE_NAME) $(GATEWAY_VERSIONED_IMAGE_NAME)
	docker tag $(ASSET_IMAGE_NAME) $(ASSET_VERSIONED_IMAGE_NAME)
	docker tag $(MIGRATE_IMAGE_NAME) $(MIGRATE_VERSIONED_IMAGE_NAME)
	docker push $(AUTH_VERSIONED_IMAGE_NAME)
	docker push $(GATEWAY_VERSIONED_IMAGE_NAME)
	docker push $(ASSET_VERSIONED_IMAGE_NAME)
	docker push $(MIGRATE_VERSIONED_IMAGE_NAME)

.PHONY: release-commit
release-commit:
	./scripts/release-commit.sh

.PHONY: preview-doc-auth
preview-doc-auth:
	./scripts/preview-doc.sh auth

.PHONY: preview-doc-asset
preview-doc-asset:
	./scripts/preview-doc.sh asset

.PHONY: generate-doc-auth
generate-doc-auth:
	@openapi3-gen -output "$(DOC_PATH)" ./cmd/auth/... ./pkg/auth/...

.PHONY: generate-doc-asset
generate-doc-asset:
	@openapi3-gen -output "$(DOC_PATH)" ./cmd/asset/... ./pkg/asset/...

.PHONY: generate-static-asset
generate-static-asset:
	cd ./scripts/deploy-asset; \
	npm ci; \
	rm -rf dist/; \
	mkdir -p "dist/git-$(GIT_SHORT_SHA)"; \
	cp -R ../../static/. "dist/git-$(GIT_SHORT_SHA)"; \
	npx postcss \
		'../../static/**/*.css' \
		--base '../../static' \
		--dir "dist/git-$(GIT_SHORT_SHA)" \
		--config postcss.config.js
