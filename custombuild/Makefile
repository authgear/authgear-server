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
GO_RUN_TAGS ::=

.PHONY: start
start:
	go run -tags "$(GO_RUN_TAGS)" -ldflags ${LDFLAGS} ./cmd/authgearx start

.PHONY: start-portal
start-portal:
	go run -tags "$(GO_RUN_TAGS)" -ldflags ${LDFLAGS} ./cmd/portalx start

.PHONY: build
build:
	go build -o $(BIN_NAME) -tags "$(GO_BUILD_TAGS)" -ldflags ${LDFLAGS} ./cmd/$(TARGET)

.PHONY: build-image
build-image:
	$(eval BUILD_OPTS ::= --tag $(IMAGE_NAME))
ifeq (${TAG_IMAGE},true) # if TAG_IMAGE
	$(eval BUILD_OPTS += --tag $(IMAGE_NAME):latest)
ifneq (${GIT_HASH},)
	$(eval BUILD_OPTS += --tag $(IMAGE_NAME):$(GIT_HASH))
endif
ifneq (${GIT_NAME},)
	$(eval BUILD_OPTS += --tag $(IMAGE_NAME):$(GIT_NAME))
endif
endif # endif TAG_IMAGE
ifeq ($(PUSH_IMAGE),true)
	$(eval BUILD_OPTS += --push)
endif
	@# Add --pull so that we are using the latest base image.
	@# The build context is the parent directory
	docker build --pull --ssh=default \
		--file ./cmd/$(TARGET)/Dockerfile \
		$(BUILD_OPTS) \
		--build-arg GIT_HASH=$(GIT_HASH) ../
