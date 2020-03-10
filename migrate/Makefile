SHELL := /bin/bash

SCHEMA ?= app_config

ifeq (1,$(WITH_CONTROLLER))
MIGRATION_ARG ?= -migration=core -migration=auth -migration=controller
else
# Sometimes only auth migration is added.
# The program complains if we run `up 1`.
#
# Intentionally not quoted
MIGRATION_ARG ?= -migration=core -migration=auth
endif

ifeq (1,$(DRY_RUN))
DRY_RUN_ARG := -dry-run
endif

AUTH_SOURCE ?= $(realpath ../migrations/auth)
CORE_SOURCE ?= $(realpath ../migrations/core)
CONTROLLER_SOURCE ?= $(realpath ../migrations/controller)

MIGRATE_CMD ?= version

ifneq (,$(APP_FILTER_KEY))
APP_FILTER_KEY_ARG := -app-filter-key=$(APP_FILTER_KEY)
endif

ifneq (,$(APP_FILTER_VALUE))
APP_FILTER_VALUE_ARG := -app-filter-value=$(APP_FILTER_VALUE)
endif

ifneq (,$(CONFIG_DATABASE))
CONFIG_DATABASE_ARG := -config-database=$(CONFIG_DATABASE)
endif

ifneq (,$(HOSTNAME_OVERRIDE))
HOSTNAME_OVERRIDE_ARG := -hostname-override=$(HOSTNAME_OVERRIDE)
endif

ifneq (,$(DATABASE_URL))
DATABASE_URL_ARG := -database=$(DATABASE_URL)
endif

.PHONY: migrate
migrate:
	go run cmd/migrate/main.go \
		-add-migration-src=auth,$(AUTH_SOURCE) \
		-add-migration-src=core,$(CORE_SOURCE) \
		-add-migration-src=controller,$(CONTROLLER_SOURCE) \
		-schema $(SCHEMA) \
		$(MIGRATION_ARG) \
		$(DRY_RUN_ARG) \
		$(APP_FILTER_KEY_ARG) \
		$(APP_FILTER_VALUE_ARG) \
		$(CONFIG_DATABASE_ARG) \
		$(DATABASE_URL_ARG) \
		$(MIGRATE_CMD)

.PHONY: http
http:
	go run cmd/migrate/main.go \
		-add-migration-src=auth,$(AUTH_SOURCE) \
		-add-migration-src=core,$(CORE_SOURCE) \
		-http-server

.PHONY: add-version
add-version: MODULE ?= INVALID
add-version: MIGRATION_DIR ?= $(realpath ../migrations/$(MODULE))
add-version:
	cp $(MIGRATION_DIR)/templates/template.down.sql $(MIGRATION_DIR)/$(shell date +%s)_${REVISION}.down.sql
	cp $(MIGRATION_DIR)/templates/template.up.sql $(MIGRATION_DIR)/$(shell date +%s)_${REVISION}.up.sql
