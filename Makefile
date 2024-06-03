# GIT_NAME could be empty.
GIT_NAME ?= $(shell git describe --exact-match 2>/dev/null)
GIT_HASH ?= git-$(shell git rev-parse --short=12 HEAD)

LDFLAGS ?= "-X github.com/authgear/authgear-server/pkg/version.Version=${GIT_HASH}"

.PHONY: start
start:
	go run -ldflags ${LDFLAGS} ./cmd/authgear start

.PHONY: start-portal
start-portal:
	go run -ldflags ${LDFLAGS} ./cmd/portal start

.PHONY: vendor
vendor:
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.55.2
	go mod download
	go install github.com/golang/mock/mockgen
	go install github.com/google/wire/cmd/wire
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install go.k6.io/xk6/cmd/xk6@latest
	npm --prefix ./scripts/npm ci
	npm --prefix ./authui ci
	npm --prefix ./portal ci
	$(MAKE) authui
	$(MAKE) portal

.PHONY: go-mod-outdated
go-mod-outdated:
	# https://stackoverflow.com/questions/55866604/whats-the-go-mod-equivalent-of-npm-outdated
	# Since go 1.21, this command will exit 2 when one of the dependencies require a go version newer than us.
	# This implies we have to use the latest verion of Go whenever possible.
	go list -u -m -f '{{if .Update}}{{if not .Indirect}}{{.}}{{end}}{{end}}' all

.PHONY: ensure-important-modules-up-to-date
ensure-important-modules-up-to-date:
	# If grep matches something, it exits 0, otherwise it exits 1.
	# In our case, we want to invert the exit code.
	$(MAKE) go-mod-outdated | grep "github.com/nyaruka/phonenumbers"; status_code=$$?; if [ $$status_code -eq 0 ]; then exit 1; else exit 0; fi;

.PHONY: generate
generate:
	go generate ./pkg/... ./cmd/... ./e2e/...

.PHONY: test
test:
	$(MAKE) -C ./k6 go-test
	go test ./pkg/... -timeout 1m30s

.PHONY: lint
lint:
	golangci-lint run ./cmd/... ./pkg/... --timeout 7m
	go run ./devtools/translationlinter
	-go run ./devtools/importlinter api api >.make-lint-expect 2>&1
	-go run ./devtools/importlinter lib api util >> .make-lint-expect 2>&1
	-go run ./devtools/importlinter admin api lib util >> .make-lint-expect 2>&1
	-go run ./devtools/importlinter auth api lib util >> .make-lint-expect 2>&1
	-go run ./devtools/importlinter portal api lib util >> .make-lint-expect 2>&1
	-go run ./devtools/importlinter resolver api lib util >> .make-lint-expect 2>&1
	-go run ./devtools/importlinter util api util >> .make-lint-expect 2>&1
	-go run ./devtools/importlinter version version >> .make-lint-expect 2>&1
	-go run ./devtools/importlinter worker api lib util >> .make-lint-expect 2>&1
	-go run ./devtools/bandimportlinter ./pkg ./cmd >> .make-lint-expect 2>&1
	git diff --exit-code .make-lint-expect > /dev/null 2>&1
	go run ./devtools/gotemplatelinter ./resources/authgear/templates/en/web/authflowv2

.PHONY: fmt
fmt:
	# Ignore generated files, such as wire_gen.go and *_mock_test.go
	find ./pkg ./cmd ./e2e -name '*.go' -not -name 'wire_gen.go' -not -name '*_mock_test.go' | sort | xargs goimports -w -format-only -local github.com/authgear/authgear-server

.PHONY: govulncheck
govulncheck:
	govulncheck -show traces,version,verbose ./...

# osusergo: https://godoc.org/github.com/golang/go/src/os/user
# netgo: https://golang.org/doc/go1.5#net
# static_build: https://github.com/golang/go/issues/26492#issuecomment-635563222
#   The binary is static on Linux only. It is not static on macOS.
# timetzdata: https://golang.org/doc/go1.15#time/tzdata
.PHONY: build
build:
	go build -o $(BIN_NAME) -tags "osusergo netgo static_build timetzdata $(GO_BUILD_TAGS)" -ldflags ${LDFLAGS} ./cmd/$(TARGET)

.PHONY: binary
binary:
	rm -rf ./dist
	mkdir ./dist
	$(MAKE) build GO_BUILD_TAGS=authgearlite TARGET=authgear BIN_NAME=./dist/authgear-lite-"$(shell go env GOOS)"-"$(shell go env GOARCH)"-${GIT_HASH}
	$(MAKE) build GO_BUILD_TAGS=authgearlite TARGET=portal BIN_NAME=./dist/authgear-portal-lite-"$(shell go env GOOS)"-"$(shell go env GOARCH)"-${GIT_HASH}

.PHONY: check-tidy
check-tidy:
	$(MAKE) fmt
	$(MAKE) generate
	$(MAKE) html-email
	$(MAKE) export-schemas
	$(MAKE) generate-timezones
	$(MAKE) generate-rtl
	$(MAKE) generate-twemoji-icons
	$(MAKE) generate-material-icons
	go mod tidy
	git status --porcelain | grep '.*'; test $$? -eq 1

.PHONY: build-image
build-image:
	# Add --pull so that we are using the latest base image.
	docker build --pull --file ./cmd/$(TARGET)/Dockerfile --tag $(IMAGE_NAME) --build-arg GIT_HASH=$(GIT_HASH) .

.PHONY: tag-image
tag-image: DOCKER_IMAGE = quay.io/theauthgear/$(IMAGE_NAME)
tag-image:
	docker tag $(IMAGE_NAME) $(DOCKER_IMAGE):latest
	docker tag $(IMAGE_NAME) $(DOCKER_IMAGE):$(GIT_HASH)
	if [ ! -z $(GIT_NAME) ]; then docker tag $(IMAGE_NAME) $(DOCKER_IMAGE):$(GIT_NAME); fi

.PHONY: push-image
push-image: DOCKER_IMAGE = quay.io/theauthgear/$(IMAGE_NAME)
push-image:
	docker manifest inspect $(DOCKER_IMAGE):$(GIT_HASH) > /dev/null; if [ $$? -eq 0 ]; then \
		echo "$(DOCKER_IMAGE):$(GIT_HASH) exists. Skip push"; \
	else \
		docker push $(DOCKER_IMAGE):latest ;\
		docker push $(DOCKER_IMAGE):$(GIT_HASH) ;\
		if [ ! -z $(GIT_NAME) ]; then docker push $(DOCKER_IMAGE):$(GIT_NAME); fi ;\
	fi

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
	rm resources/authgear/generated/build*.html

.PHONY: portal
portal:
	npm run --silent --prefix ./portal build
	cp -R ./portal/dist/ ./resources/portal/static/

# After you run `make clean`, you have to run `make authui` and `make portal`.
.PHONY: clean
clean:
	rm -rf ./resources/portal/static
	git checkout -- ./resources/portal/static
	# It is important NOT to remove the directory.
	# Otherwise the watcher is stopped.
	rm -rf ./resources/authgear/generated/*
	git checkout -- ./resources/authgear/generated/*
	rm -rf ./portal/.parcel-cache

.PHONY: export-schemas
export-schemas:
	go run ./scripts/exportschemas -s app-config -o tmp/app-config.schema.json
	go run ./scripts/exportschemas -s secrets-config -o tmp/secrets-config.schema.json
	npm run --silent --prefix ./scripts/npm export-graphql-schema admin > portal/src/graphql/adminapi/schema.graphql
	npm run --silent --prefix ./scripts/npm export-graphql-schema portal > portal/src/graphql/portal/schema.graphql

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

.PHONY: generate-material-icons
generate-material-icons:
	make -C ./scripts/python generate-material-icons

.PHONY: generate-twemoji-icons
generate-twemoji-icons:
	make -C ./scripts/python generate-twemoji-icons

.PHONY: logs-summary
logs-summary:
	git log --first-parent --format='%as (%h) %s' $(A)..$(B)

.PHONY: mkcert
mkcert:
	rm -f tls-cert.pem tls-key.pem
	mkcert -cert-file tls-cert.pem -key-file tls-key.pem "::1" "127.0.0.1" localhost portal.localhost accounts.localhost accounts.portal.localhost $$(ifconfig | grep 'inet 192' | awk '{print $$2}') $$(ifconfig | grep 'inet 192' | awk '{print $$2}').nip.io

.PHONY: check-dockerignore
check-dockerignore:
	./scripts/sh/check-dockerignore.sh
