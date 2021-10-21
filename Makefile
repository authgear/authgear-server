# GIT_NAME could be empty.
GIT_NAME ?= $(shell git describe --exact-match 2>/dev/null)
GIT_HASH ?= git-$(shell git rev-parse --short=12 HEAD)

LDFLAGS ?= "-X github.com/authgear/authgear-server/pkg/version.Version=${GIT_HASH}"

.PHONY: start
start:
	go run -ldflags ${LDFLAGS} ./cmd/authgear start

.PHONY: vendor
vendor:
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.41.1
	go mod download
	go install github.com/golang/mock/mockgen
	go install github.com/google/wire/cmd/wire
	(cd scripts/npm && npm ci)
	(cd authui && npm ci)
	$(MAKE) authui

.PHONY: go-mod-outdated
go-mod-outdated:
	# https://stackoverflow.com/questions/55866604/whats-the-go-mod-equivalent-of-npm-outdated
	go list -u -m -f '{{if .Update}}{{if not .Indirect}}{{.}}{{end}}{{end}}' all

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
	-go run ./devtools/translationlinter
	-go run ./devtools/bandimportlinter ./pkg ./cmd >> .make-lint-expect 2>&1
	git diff --exit-code .make-lint-expect > /dev/null 2>&1

.PHONY: fmt
fmt:
	go fmt ./...

# osusergo: https://godoc.org/github.com/golang/go/src/os/user
# netgo: https://golang.org/doc/go1.5#net
# static_build: https://github.com/golang/go/issues/26492#issuecomment-635563222
#   The binary is static on Linux only. It is not static on macOS.
# timetzdata: https://golang.org/doc/go1.15#time/tzdata
.PHONY: build
build:
	go build -o $(BIN_NAME) -tags 'osusergo netgo static_build timetzdata' -ldflags ${LDFLAGS} ./cmd/$(TARGET)

.PHONY: check-tidy
check-tidy:
	$(MAKE) fmt
	$(MAKE) generate
	$(MAKE) html-email
	$(MAKE) export-schemas
	$(MAKE) generate-timezones
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
	docker push $(DOCKER_IMAGE):latest
	docker push $(DOCKER_IMAGE):$(GIT_HASH)
	if [ ! -z $(GIT_NAME) ]; then docker push $(DOCKER_IMAGE):$(GIT_NAME); fi

.PHONY: html-email
html-email:
	for t in $$(find resources -name '*.mjml'); do \
		./scripts/npm/node_modules/.bin/mjml -l strict "$$t" > "$${t%.mjml}.html"; \
	done

.PHONY: authui
authui:
	# Build Auth UI
	npm run --silent --prefix ./authui typecheck
	npm run --silent --prefix ./authui format
	npm run --silent --prefix ./authui build

.PHONY: portal
portal:
	npm run --silent --prefix ./portal build
	cp -R ./portal/dist/ ./resources/portal/static/

# After you run `make clean`, you have to run `make authui` and `make portal`.
.PHONY: clean
clean:
	rm -rf ./resources/portal/static
	git checkout -- ./resources/portal/static
	rm -rf ./resources/authgear/static
	git checkout -- ./resources/authgear/static

.PHONY: export-schemas
export-schemas:
	go run ./scripts/exportschemas -s app-config -o tmp/app-config.schema.json
	go run ./scripts/exportschemas -s secrets-config -o tmp/secrets-config.schema.json
	npm run --silent --prefix ./scripts/npm export-graphql-schema admin > portal/src/graphql/adminapi/schema.graphql
	npm run --silent --prefix ./scripts/npm export-graphql-schema portal > portal/src/graphql/portal/schema.graphql

.PHONY:	generate-timezones
generate-timezones:
	npm run --silent --prefix ./scripts/npm generate-go-timezones > pkg/util/tzutil/names.go

.PHONY: logs-summary
logs-summary:
	git log --first-parent --format='%as (%h) %s' $(A)..$(B)
