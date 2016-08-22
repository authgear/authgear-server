GIT_SHA := $(shell git rev-parse HEAD)
GIT_SHORT_SHA := $(shell git rev-parse --short HEAD)
VERSION := $(shell git describe --always --tags)
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

DOCKER_REGISTRY := 
DOCKER_ORG_NAME := skygeario
DOCKER_IMAGE := skygear-server-deps
DOCKER_TAG := git-$(shell git rev-parse --short HEAD)
PUSH_DOCKER_TAG := $(VERSION)

IMAGE_NAME := $(DOCKER_REGISTRY)$(DOCKER_ORG_NAME)/$(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: clean
clean:
	rm skygear-server

skygear-server:
	$(MAKE) -C ../../../ vendor WITH_DOCKER=1
	$(MAKE) -C ../../../ build WITH_DOCKER=1 WITH_ZMQ=1 DIST=dist/skygear-server
	cp ../../../dist/skygear-server .

.PHONY: docker-build
docker-build: skygear-server
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

