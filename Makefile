# GIT_NAME could be empty.
GIT_NAME ?= $(shell git describe --exact-match 2>/dev/null)
GIT_HASH ?= git-$(shell git rev-parse --short=12 HEAD)
LDFLAGS ?= "-X github.com/authgear/authgear-server/pkg/version.Version=${GIT_HASH}"
DOCKER_TEMP_TAG ?= authgear-server
DOCKER_IMAGE ?= quay.io/theauthgear/authgear-server

.PHONY: start
start:
	go run -ldflags ${LDFLAGS} ./cmd/authgear start

.PHONY: vendor
vendor:
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.27.0
	go mod download
	go install github.com/golang/mock/mockgen
	go install github.com/google/wire/cmd/wire
	(cd scripts/npm && npm ci)

.PHONY: generate
generate:
	go generate ./pkg/... ./cmd/...

.PHONY: test
test:
	go test ./pkg/... -timeout 1m30s

.PHONY: lint
lint:
	golangci-lint run ./cmd/... ./pkg/...
	-go run ./devtools/importlinter api api >.make-lint-expect 2>&1
	-go run ./devtools/importlinter lib api util >> .make-lint-expect 2>&1
	-go run ./devtools/importlinter admin api lib util >> .make-lint-expect 2>&1
	-go run ./devtools/importlinter auth api lib util >> .make-lint-expect 2>&1
	-go run ./devtools/importlinter portal api lib util >> .make-lint-expect 2>&1
	-go run ./devtools/importlinter resolver api lib util >> .make-lint-expect 2>&1
	-go run ./devtools/importlinter util api util >> .make-lint-expect 2>&1
	-go run ./devtools/importlinter version version >> .make-lint-expect 2>&1
	-go run ./devtools/importlinter worker api lib util >> .make-lint-expect 2>&1
	git diff --exit-code .make-lint-expect > /dev/null 2>&1

.PHONY: fmt
fmt:
	go fmt ./...

# The -tags builds static binary on linux.
# On macOS the binary is NOT static though.
# https://github.com/golang/go/issues/26492#issuecomment-635563222
.PHONY: build
build:
	go build -o authgear -tags 'osusergo netgo static_build' -ldflags ${LDFLAGS} ./cmd/authgear

.PHONY: check-tidy
check-tidy:
	$(MAKE) generate
	$(MAKE) html-email
	go mod tidy
	git status --porcelain | grep '.*'; test $$? -eq 1

.PHONY: build-image
build-image:
	docker build --tag $(DOCKER_TEMP_TAG) --build-arg GIT_HASH=$(GIT_HASH) .

.PHONY: tag-image
tag-image:
	docker tag $(DOCKER_TEMP_TAG) $(DOCKER_IMAGE):latest
	docker tag $(DOCKER_TEMP_TAG) $(DOCKER_IMAGE):$(GIT_HASH)
	if [ ! -z $(GIT_NAME) ]; then docker tag $(DOCKER_TEMP_TAG) $(DOCKER_IMAGE):$(GIT_NAME); fi

.PHONY: push-image
push-image:
	docker push $(DOCKER_IMAGE):latest
	docker push $(DOCKER_IMAGE):$(GIT_HASH)
	if [ ! -z $(GIT_NAME) ]; then docker push $(DOCKER_IMAGE):$(GIT_NAME); fi

.PHONY: html-email
html-email:
	for t in templates/*.mjml; do \
		./scripts/npm/node_modules/.bin/mjml -l strict "$$t" > "$${t%.mjml}.html"; \
	done

.PHONY: static
static:
	rm -rf ./dist
	mkdir -p "./dist/${GIT_HASH}"
	# Start by copying src
	cp -R ./static/. "./dist/${GIT_HASH}"
	# Process CSS
	./scripts/npm/node_modules/.bin/postcss './static/**/*.css' --base './static/' --dir "./dist/${GIT_HASH}" --config ./scripts/npm/postcss.config.js

.PHONY: export-schemas
export-schemas:
	go run ./scripts/exportschemas -s app-config -o tmp/app-config.schema.json
	go run ./scripts/exportschemas -s secrets-config -o tmp/secrets-config.schema.json
	npm run --silent --prefix ./scripts/npm export-graphql-schema admin > tmp/admin-schema.graphql
	npm run --silent --prefix ./scripts/npm export-graphql-schema portal > tmp/portal-schema.graphql