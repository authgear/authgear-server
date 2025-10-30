CMD_AUTHGEAR ::= authgear
CMD_PORTAL ::= portal
BUILD_CTX ::= .

include ./makefiles/common.mk
include ./makefiles/go-mod-outdated.mk
include ./makefiles/govulncheck.mk

.PHONY: authgearonce-start
authgearonce-start: GO_RUN_TAGS += authgearonce
authgearonce-start:
	$(MAKE) start GO_RUN_TAGS::="$(GO_RUN_TAGS)"

.PHONY: authgearonce-start-portal
authgearonce-start-portal: GO_RUN_TAGS += authgearonce
authgearonce-start-portal:
	$(MAKE) start-portal GO_RUN_TAGS::="$(GO_RUN_TAGS)"

.PHONY: vendor
vendor:
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin v2.5.0
	go mod download
	$(MAKE) build-frondend

.PHONY: build-frondend
build-frondend:
	npm --prefix ./scripts/npm ci
	npm --prefix ./authui ci
	npm --prefix ./portal ci
	$(MAKE) authui
	$(MAKE) portal

# This makefile target automates running `go mod tidy` in every directory that has a go.mod file.
.PHONY: go-mod-tidy
go-mod-tidy:
	find . -name 'go.mod' -exec sh -c 'cd $$(dirname {}); go mod tidy' \;

.PHONY: ensure-important-modules-up-to-date
ensure-important-modules-up-to-date:
	./scripts/sh/ensure-important-modules-up-to-date.sh

.PHONY: generate
generate:
	go generate ./pkg/... ./cmd/...
	cd ./e2e && go generate ./...

.PHONY: test
test:
	$(MAKE) -C ./k6 go-test
	go test ./devtools/goanalysis/... ./cmd/... ./pkg/... -timeout 1m30s

.PHONY: lint-translation-keys
lint-translation-keys:
	-go run ./devtools/gotemplatelinter --ignore-rule indentation --ignore-rule eol-at-eof ./resources/authgear/templates/en/web/authflowv2 >.make-lint-translation-keys-expect 2>&1
	git diff --exit-code .make-lint-translation-keys-expect > /dev/null 2>&1

.PHONY: lint
lint:
	golangci-lint run ./cmd/... ./pkg/... --timeout 7m
	go run ./devtools/translationlinter
	-go run ./devtools/importlinter api api >.make-lint-expect 2>&1
	-go run ./devtools/importlinter lib api util >> .make-lint-expect 2>&1
	-go run ./devtools/importlinter admin api lib util graphqlgo >> .make-lint-expect 2>&1
	-go run ./devtools/importlinter auth api lib util >> .make-lint-expect 2>&1
	-go run ./devtools/importlinter portal api lib util >> .make-lint-expect 2>&1
	-go run ./devtools/importlinter resolver api lib util >> .make-lint-expect 2>&1
	-go run ./devtools/importlinter util api util >> .make-lint-expect 2>&1
	-go run ./devtools/importlinter version version >> .make-lint-expect 2>&1
	-go run ./devtools/importlinter worker api lib util >> .make-lint-expect 2>&1
	-go run ./devtools/bandimportlinter ./pkg ./cmd >> .make-lint-expect 2>&1
	git diff --exit-code .make-lint-expect > /dev/null 2>&1
	go run ./devtools/gotemplatelinter --ignore-rule translation-key ./resources/authgear/templates/en/web/authflowv2
	$(MAKE) lint-translation-keys
	go run ./devtools/goanalysis ./cmd/... ./pkg/...

.PHONY: sort-translations
sort-translations:
	go run ./devtools/translationsorter

.PHONY: sort-vettedpositions
sort-vettedpositions:
	./scripts/python/sort_vettedpositions.py ./.vettedpositions

.PHONY: fmt
fmt:
	# Ignore generated files, such as wire_gen.go and *_mock_test.go
	find ./devtools ./pkg ./cmd ./e2e -name '*.go' -not -name '*_gen.go' -not -name '*_mock_test.go' | sort | xargs go tool goimports -w -format-only -local github.com/authgear/authgear-server
	$(MAKE) sort-translations

.PHONY: binary
binary:
	rm -rf ./dist
	mkdir ./dist
	$(MAKE) build AUTHGEARLITE::=1 TARGET::=authgear BIN_NAME::=./dist/authgear-lite-"$(shell go env GOOS)"-"$(shell go env GOARCH)"-${GIT_HASH}
	$(MAKE) build AUTHGEARLITE::=1 TARGET::=portal BIN_NAME::=./dist/authgear-portal-lite-"$(shell go env GOOS)"-"$(shell go env GOARCH)"-${GIT_HASH}

.PHONY: check-tidy
check-tidy:
	# For some unknown reason, `make generate` will somehow format the files again (but with a different rule).
	# So `make fmt` has to be run after `make generate`.
	$(MAKE) sort-vettedpositions
	$(MAKE) generate
	$(MAKE) fmt
	$(MAKE) html-email
	$(MAKE) export-schemas
	$(MAKE) generate-timezones
	$(MAKE) generate-rtl
	$(MAKE) generate-twemoji-icons
	$(MAKE) generate-material-icons
	$(MAKE) graphiql
	$(MAKE) once/Dockerfile
	go mod tidy
	# We wanted to run the following, but that requires SSH, which does not work for running CI for PRs.
	# (cd custombuild && go mod tidy)
	git status --porcelain | grep '.*'; test $$? -eq 1

	$(MAKE) -C authui check-tidy
	$(MAKE) -C portal check-tidy

.PHONY: html-email
html-email:
	# Generate `.mjml` templates from `.mjml.gotemplate` files
	go run ./scripts/generatemjml/main.go -i resources/authgear/templates

	for t in $$(find resources -name '*.mjml'); do \
		./scripts/npm/node_modules/.bin/mjml -l strict "$$t" > "$${t%.mjml}.html"; \
	done

.PHONY: authui
authui:
	# Build Auth UI
	npm run --silent --prefix ./authui typecheck
	npm run --silent --prefix ./authui format
	npm run --silent --prefix ./authui build
	# Vite by default will remove the output directory before the build.
	# So we need to touch .gitkeep after the build to avoid git changes.
	touch resources/authgear/generated/.gitkeep

.PHONY: authui-dev
authui-dev:
	# Make sure that assets are generated before starting dev server
	$(MAKE) authui
	# Start development server for Auth UI
	npm run --silent --prefix ./authui dev

.PHONY: portal
portal:
	npm run --silent --prefix ./portal build
	cp -R ./portal/dist/. ./resources/portal/static/
	# Vite by default will remove the output directory before the build.
	# So we need to touch .gitkeep after the build to avoid git changes.
	touch ./portal/dist/.gitkeep

# After you run `make clean`, you have to run `make authui` and `make portal`.
.PHONY: clean
clean:
	rm -rf ./resources/portal/static
	git checkout -- ./resources/portal/static
	# It is important NOT to remove the directory.
	# Otherwise the watcher is stopped.
	rm -rf ./resources/authgear/generated/*
	git checkout -- ./resources/authgear/generated/*
	rm -rf ./portal/dist
	git checkout -- ./portal/dist
	rm -rf ./e2e/logs
	git checkout -- ./e2e/logs
	find . -name '.parcel-cache' -exec rm -rf '{}' \;
	find . -name '.stylelintcache' -exec rm -rf '{}' \;
	$(MAKE) -C ./e2e clean
	$(MAKE) -C ./k6 clean

.PHONY: export-schemas
export-schemas:
	go run ./scripts/exportschemas -s app-config -o tmp/app-config.schema.json
	go run ./scripts/exportschemas -s secrets-config -o tmp/secrets-config.schema.json
	npm run --silent --prefix ./scripts/npm export-graphql-schema admin > portal/src/graphql/adminapi/schema.graphql
	npm run --silent --prefix ./scripts/npm export-graphql-schema portal > portal/src/graphql/portal/schema.graphql
	cd portal && npm run gentype

.PHONY: export-v2-translations
export-v2-translations:
	@npm run --silent --prefix ./scripts/npm export-v2-translations

.PHONY: import-v2-translations
import-v2-translations:
	@npm run --silent --prefix ./scripts/npm import-v2-translations

.PHONY:	generate-timezones
generate-timezones:
	npm run --silent --prefix ./scripts/npm generate-go-timezones > pkg/util/tzutil/names.go

.PHONY: generate-rtl
generate-rtl:
	go run ./scripts/characterorder/main.go | gofmt > pkg/util/intl/rtl_map.go

.PHONY: check-spell-translations
check-spell-translations:
	@npm run --prefix ./scripts/npm check-spell-translations

.PHONY: generate-material-icons
generate-material-icons:
	$(MAKE) -C ./scripts/python generate-material-icons

.PHONY: generate-twemoji-icons
generate-twemoji-icons:
	$(MAKE) -C ./scripts/python generate-twemoji-icons

# This make target helps you in updating an existing translation key.
# When you are asked to update an update existing translation key, you do
# 1. Update the value in English.
# 2. `make translation-json-del-key KEY=the-key` to remove the key in other translation JSON files.
# 3. `cd scripts/python; make generate-translations` to re-generate the missing key.
.PHONY: translation-json-del-key
translation-json-del-key: KEY=
translation-json-del-key:
	find . -path './resources/authgear/templates/*/translation.json' -not -path './resources/authgear/templates/*/messages/translation.json' -not -path './resources/authgear/templates/en/translation.json' -exec sh -c "jq <\$$1 'del(.[\"$(KEY)\"])' > \$$1.tmp; mv \$$1.tmp \$$1" _ '{}' \;

.PHONY: logs-summary
logs-summary:
	git log --first-parent --format='%as (%h) %s' $(A)..$(B)

.PHONY: mkcert-ca
mkcert-ca:
	cp "$$(mkcert -CAROOT)"/rootCA.pem ./rootCA.pem

.PHONY: mkcert
mkcert:
	rm -f tls-cert.pem tls-key.pem
	mkcert -cert-file tls-cert.pem -key-file tls-key.pem "::1" "127.0.0.1" postgres16 localhost portal.localhost accounts.localhost accounts.portal.localhost $$(ifconfig | grep 'inet 192' | awk '{print $$2}') $$(ifconfig | grep 'inet 192' | awk '{print $$2}').nip.io

.PHONY: check-dockerignore
check-dockerignore:
	./scripts/sh/check-dockerignore.sh

.PHONY: graphiql
graphiql:
	npm --prefix portalgraphiql ci
	npm --prefix portalgraphiql run build
	cp ./portalgraphiql/dist/index.html pkg/util/graphqlutil/graphiql.html

.PHONY: once/Dockerfile
once/Dockerfile:
	rm -f $@
	touch $@
	cat ./once/opening.dockerfile >> $@
	printf "\n" >> $@
	sed -e '/^# syntax=/d' ./cmd/authgear/Dockerfile >> $@
	printf "\n" >> $@
	sed -e '/^# syntax=/d' ./cmd/portal/Dockerfile >> $@
	printf "\n" >> $@
	sed -e '/^# syntax=/d' ./once/partial.dockerfile >> $@

.PHONY: authgearonce-set-git-tag-name
authgearonce-set-git-tag-name:
	@./scripts/sh/authgearonce-set-git-tag-name.sh

.PHONY: authgearonce-set-AUTHGEARONCE_LICENSE_SERVER_ENV-by-tag-name
authgearonce-set-AUTHGEARONCE_LICENSE_SERVER_ENV-by-tag-name:
	@./scripts/sh/authgearonce-set-AUTHGEARONCE_LICENSE_SERVER_ENV-by-tag-name.sh

.PHONY: authgearonce-binary
ifeq ($(AUTHGEARONCE_LICENSE_SERVER_ENV), local)
GO_BUILD_TAGS += authgearonce_license_server_local
endif
ifeq ($(AUTHGEARONCE_LICENSE_SERVER_ENV), staging)
GO_BUILD_TAGS += authgearonce_license_server_staging
endif
authgearonce-binary:
	rm -rf ./dist
	mkdir ./dist
	$(MAKE) build TARGET::=once BIN_NAME::=./dist/authgear-once-"$(shell go env GOOS)"-"$(shell go env GOARCH)"
