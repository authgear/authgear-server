ifeq ($(origin GIT_HASH), undefined)
GIT_HASH ::= git-$(shell git rev-parse --short=12 HEAD)
endif
ifeq ($(origin LDFLAGS), undefined)
LDFLAGS ::= "-X github.com/authgear/authgear-server/pkg/version.Version=${GIT_HASH}"
endif


# osusergo: https://godoc.org/github.com/golang/go/src/os/user
# netgo: https://golang.org/doc/go1.5#net
# static_build: https://github.com/golang/go/issues/26492#issuecomment-635563222
#   The binary is static on Linux only. It is not static on macOS.
# timetzdata: https://golang.org/doc/go1.15#time/tzdata
GO_BUILD_TAGS ::= osusergo netgo static_build timetzdata
# Since it is not possible to write `make GO_BUILD_TAGS+=authgearonce` in the command line,
# we use a simpler mechanism to append to GO_BUILD_TAGS.
# We define AUTHGEARLITE and AUTHGEARONCE, and if they are 1, then the corresponding build tag is appended.
ifeq ($(AUTHGEARLITE), 1)
GO_BUILD_TAGS += authgearlite
endif
ifeq ($(AUTHGEARONCE), 1)
GO_BUILD_TAGS += authgearonce
endif
# authgeardev: This build tag represents the build is for local development purpose.
#              Currently, it affects whether the builtin resource FS uses OS FS or embed.FS.
#              See ./pkg/util/resource/manager.go for details.
GO_RUN_TAGS ::= authgeardev


.PHONY: start
start:
	go run -tags "$(GO_RUN_TAGS)" -ldflags ${LDFLAGS} ./cmd/${CMD_AUTHGEAR} start

.PHONY: start-background
start-background:
	go run -tags "$(GO_RUN_TAGS)" -ldflags ${LDFLAGS} ./cmd/${CMD_AUTHGEAR} background

.PHONY: start-portal
start-portal:
	go run -tags "$(GO_RUN_TAGS)" -ldflags ${LDFLAGS} ./cmd/${CMD_PORTAL} start

.PHONY: build
build:
	go build -o $(BIN_NAME) -tags "$(GO_BUILD_TAGS)" -ldflags ${LDFLAGS} ./cmd/$(TARGET)


.PHONY: build-image
ifeq ($(origin DOCKERFILE), undefined)
build-image: DOCKERFILE ::= ./Dockerfile
endif

build-image: BUILD_OPTS ::=
ifeq ($(BUILD_ARCH),amd64)
build-image: BUILD_OPTS += --platform linux/$(BUILD_ARCH)
else ifeq ($(BUILD_ARCH),arm64)
build-image: BUILD_OPTS += --platform linux/$(BUILD_ARCH)
endif
ifneq ($(OUTPUT),)
build-image: BUILD_OPTS += --output=$(OUTPUT)
endif
ifneq ($(EXTRA_BUILD_OPTS),)
build-image: BUILD_OPTS += $(EXTRA_BUILD_OPTS)
endif
ifneq ($(METADATA_FILE),)
build-image: BUILD_OPTS += --metadata-file $(METADATA_FILE)
endif

# Add --pull so that we are using the latest base image.
# The build context is the parent directory
# --provenance=false because we have no idea to figure out how to deal with the unknown manifest yet.
# See https://github.com/authgear/authgear-server/pull/4943#discussion_r1891263998
build-image:
	docker build --pull \
		--provenance=false \
		--file "$(DOCKERFILE)" \
		$(BUILD_OPTS) \
		--build-arg GIT_HASH=$(GIT_HASH) ${BUILD_CTX}

.PHONY: tag-image
tag-image: TAGS ::= --tag $(IMAGE_NAME):$(GIT_HASH)
ifneq (${GIT_TAG_NAME},)
tag-image: TAGS += --tag $(IMAGE_NAME):$(GIT_TAG_NAME)
tag-image: TAGS += --tag $(IMAGE_NAME):release-$(GIT_HASH)
tag-image: TAGS += --tag $(IMAGE_NAME):release-$(GIT_TAG_NAME)
endif

tag-image: IMAGE_SOURCES ::=
tag-image: IMAGE_SOURCES := $(foreach digest,$(SOURCE_DIGESTS),${IMAGE_NAME}@${digest} )
tag-image:
	docker buildx imagetools create \
		$(TAGS) \
		${IMAGE_SOURCES}
