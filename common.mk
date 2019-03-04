DIST ?= skygear

DOCKER_COMPOSE_CMD := docker-compose -f docker-compose.make.yml

ifeq (1,${WITH_DOCKER})	
DOCKER_RUN := ${DOCKER_COMPOSE_CMD} run --rm app
endif

.PHONY: build
build:
	$(DOCKER_RUN) go build -o $(DIST)
	$(DOCKER_RUN) chmod +x $(DIST)

.PHONY: clean
clean:
	-rm $(DIST)

