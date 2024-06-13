# GIT_NAME could be empty.
GIT_NAME ?= $(shell git describe --exact-match 2>/dev/null)
GIT_HASH ?= git-$(shell git rev-parse --short=12 HEAD)

LDFLAGS ?= "-X github.com/authgear/authgear-server/pkg/version.Version=${GIT_HASH}"

.PHONY: start
start:
	go run -ldflags ${LDFLAGS} ./cmd/authgearx start

.PHONY: start-portal
start-portal:
	go run -ldflags ${LDFLAGS} ./cmd/portalx start

.PHONY: build
build:
	go build -o $(BIN_NAME) -tags "osusergo netgo static_build timetzdata $(GO_BUILD_TAGS)" -ldflags ${LDFLAGS} ./cmd/$(TARGET)

.PHONY: build-image
build-image:
	# Add --pull so that we are using the latest base image.
	# The build context is the parent directory
	docker build --pull --ssh=default --file ./cmd/$(TARGET)/Dockerfile --tag $(IMAGE_NAME) --build-arg GIT_HASH=$(GIT_HASH) ../

.PHONY: tag-image
tag-image:
	docker tag $(IMAGE_NAME) $(REMOTE_IMAGE_NAME):$(GIT_HASH)

.PHONY: push-image
push-image:
	docker push $(REMOTE_IMAGE_NAME):$(GIT_HASH)
