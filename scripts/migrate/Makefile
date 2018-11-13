MIGRATION_DIR ?= cmd/migrate/revisions/${MODULE}
SHELL := /bin/bash

.PHONY: add-version
add-version:
	cp $(MIGRATION_DIR)/templates/template.down.sql $(MIGRATION_DIR)/$(shell date +%s)_${REVISION}.down.sql
	cp $(MIGRATION_DIR)/templates/template.up.sql $(MIGRATION_DIR)/$(shell date +%s)_${REVISION}.up.sql
