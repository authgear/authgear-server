DOCKER_REGISTRY := 
DOCKER_ORG_NAME := skygeario
DOCKER_IMAGE := skygear-server-deps
DOCKER_TAG := latest

IMAGE_NAME := $(DOCKER_REGISTRY)$(DOCKER_ORG_NAME)/$(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE_NAME) .

.PHONY: docker-push
docker-push:
	docker push $(IMAGE_NAME)
