MIGRATION_DIR := $(shell pwd)/migrations
MIGRATION_DB ?= postgres://postgres:@localhost/postgres?sslmode=disable
SHELL := /bin/bash

MIGRATION_PATH := /migrations
MIGRATION_DOCKER_RUN := docker run --rm --network host \
	-v $(MIGRATION_DIR):$(MIGRATION_PATH) \
	migrate/migrate:v3.5.1

MIGRATION_RUN := $(MIGRATION_DOCKER_RUN)
ifeq (1,${NO_DOCKER})
MIGRATION_RUN := migrate
MIGRATION_PATH := ./migrations
endif

.PHONY: add-version
add-version:
	cp $(MIGRATION_DIR)/templates/template.down.sql $(MIGRATION_DIR)/$(shell date +%s)_${REVISION}.down.sql
	cp $(MIGRATION_DIR)/templates/template.up.sql $(MIGRATION_DIR)/$(shell date +%s)_${REVISION}.up.sql

.PHONY: migrate-up
migrate-up:
	$(MIGRATION_RUN) -path=$(MIGRATION_PATH) -database $(MIGRATION_DB) up

.PHONY: migrate-down
migrate-down:
	$(MIGRATION_RUN) -path=$(MIGRATION_PATH) -database $(MIGRATION_DB) down

.PHONY: migrate-version
migrate-version:
	$(MIGRATION_RUN) -path=$(MIGRATION_PATH) -database $(MIGRATION_DB) version
