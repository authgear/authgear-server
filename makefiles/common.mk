# The use of variables
#
# We use simply expanded variables in this Makefile.
#
# This means
# 1. You use ::= instead of = because = defines a recursively expanded variable.
#    See https://www.gnu.org/software/make/manual/html_node/Simple-Assignment.html
# 2. You use ::= instead of := because ::= is a POSIX standard.
#    See https://www.gnu.org/software/make/manual/html_node/Simple-Assignment.html
# 3. You do not use ?= because it is shorthand to define a recursively expanded variable.
#    See https://www.gnu.org/software/make/manual/html_node/Conditional-Assignment.html
#    You should use the long form documented in the above link instead.
# 4. When you override a variable in the command line, as documented in https://www.gnu.org/software/make/manual/html_node/Overriding.html
#    you specify the variable with ::= instead of = or :=
#    If you fail to do so, the variable becomes recursively expanded variable accidentally.
#
# GIT_NAME could be empty.
ifeq ($(origin GIT_NAME), undefined)
	GIT_NAME ::= $(shell git describe --exact-match 2>/dev/null)
endif
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
# authgeardev: This build tag represents the build is for local development purpose.
#              Currently, it affects whether the builtin resource FS uses OS FS or embed.FS.
#              See ./pkg/util/resource/manager.go for details.
GO_RUN_TAGS ::= authgeardev


.PHONY: start
start:
	go run -tags "$(GO_RUN_TAGS)" -ldflags ${LDFLAGS} ./cmd/${CMD_AUTHGEAR} start

.PHONY: start-portal
start-portal:
	go run -tags "$(GO_RUN_TAGS)" -ldflags ${LDFLAGS} ./cmd/${CMD_PORTAL} start

.PHONY: build
build:
	go build -o $(BIN_NAME) -tags "$(GO_BUILD_TAGS)" -ldflags ${LDFLAGS} ./cmd/$(TARGET)



.PHONY: build-image
build-image:
IMAGE_TAG_BASE ::= $(IMAGE_NAME):$(GIT_HASH)
BUILD_OPTS ::=
ifeq ($(BUILD_ARCH),amd64)
BUILD_OPTS += --platform linux/$(BUILD_ARCH)
else ifeq ($(BUILD_ARCH),arm64)
BUILD_OPTS += --platform linux/$(BUILD_ARCH)
endif
ifneq ($(OUTPUT),)
BUILD_OPTS += --output=$(OUTPUT)
endif
ifneq ($(EXTRA_BUILD_OPTS),)
BUILD_OPTS += $(EXTRA_BUILD_OPTS)
endif
ifneq ($(METADATA_FILE),)
BUILD_OPTS += --metadata-file $(METADATA_FILE)
endif
build-image:
	@# Add --pull so that we are using the latest base image.
	@# The build context is the parent directory
	@# --provenance=false because we have no idea to figure out how to deal with the unknown manifest yet.
	@# See https://github.com/authgear/authgear-server/pull/4943#discussion_r1891263998
	docker build --pull \
		--provenance=false \
		--file ./cmd/$(TARGET)/Dockerfile \
		$(BUILD_OPTS) \
		--build-arg GIT_HASH=$(GIT_HASH) ${BUILD_CTX}

.PHONY: tag-image
tag-image:
IMAGE_SOURCES ::=
TAGS ::= --tag $(IMAGE_NAME):latest
TAGS += --tag $(IMAGE_NAME):$(GIT_HASH)
ifneq (${GIT_NAME},)
TAGS += --tag $(IMAGE_NAME):$(GIT_NAME)
endif
IMAGE_SOURCES := $(foreach digest,$(SOURCE_DIGESTS),${IMAGE_NAME}@${digest} )
tag-image:
	docker buildx imagetools create \
		$(TAGS) \
		${IMAGE_SOURCES}
