DOCKER_REGISTRY ?=
DOCKER_ORG_NAME ?= skygeario
DOCKER_IMAGE ?= skygear-godev
DOCKER_TAG ?= git-$(shell git rev-parse --short HEAD)

IMAGE_NAME := $(DOCKER_ORG_NAME)/$(DOCKER_IMAGE)
PUSH_IMAGE_NAME := $(DOCKER_ORG_NAME)/$(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: all
all:
	$(MAKE) docker-build
	$(MAKE) docker-push
	$(MAKE) docker-push DOCKER_TAG=go1.8
	$(MAKE) docker-push DOCKER_TAG=latest

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE_NAME) .

.PHONY: docker-push
docker-push:
	docker tag $(IMAGE_NAME) $(DOCKER_REGISTRY)$(PUSH_IMAGE_NAME)
	docker push $(DOCKER_REGISTRY)$(PUSH_IMAGE_NAME)
