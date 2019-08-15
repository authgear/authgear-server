SHELL := /bin/bash

SCHEMA ?= app_config

ifeq (1,$(DRY_RUN))
DRY_RUN_ARG := -dry-run
endif

AUTH_SOURCE ?= $(realpath ../migrations/auth)
CORE_SOURCE ?= $(realpath ../migrations/core)

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

.PHONY: migrate
migrate:
	go run cmd/migrate/main.go \
		-add-migration-src=auth,$(AUTH_SOURCE) \
		-add-migration-src=core,$(CORE_SOURCE) \
		-migration=core \
		-migration=auth \
		-schema $(SCHEMA) \
		$(DRY_RUN_ARG) \
		$(APP_FILTER_KEY_ARG) \
		$(APP_FILTER_VALUE_ARG) \
		$(CONFIG_DATABASE_ARG) \
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
