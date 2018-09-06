MIGRATION_DIR :=  $(shell pwd)/migrations
MIGRATION_DB ?= postgres://postgres:@localhost/postgres?sslmode=disable
SHELL := /bin/bash

MIGRATION_DOCKER_RUN := docker run --rm --network host \
	-v $(MIGRATION_DIR):/migrations \
	migrate/migrate:v3.5.1

.PHONY: add-version
add-version:
	cp $(MIGRATION_DIR)/templates/template.down.sql $(MIGRATION_DIR)/$(shell date +%s)_${REVISION}.down.sql
	cp $(MIGRATION_DIR)/templates/template.up.sql $(MIGRATION_DIR)/$(shell date +%s)_${REVISION}.up.sql

.PHONY: migrate-up
migrate-up:
	$(MIGRATION_DOCKER_RUN) \
	-path=/migrations/ -database $(MIGRATION_DB) up

.PHONY: migrate-version
migrate-version:
	$(MIGRATION_DOCKER_RUN) \
	-path=/migrations/ -database $(MIGRATION_DB) version
