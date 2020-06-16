.PHONY: vendor
vendor:
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.27.0
	go mod download
	go install golang.org/x/tools/cmd/stringer
	go install github.com/tinylib/msgp
	go install github.com/golang/mock/mockgen
	go install github.com/google/wire/cmd/wire
	go install github.com/skygeario/openapi3-gen/cmd/openapi3-gen
	$(MAKE) -C migrate build

.PHONY: generate
generate:
	go generate ./pkg/...

.PHONY: test
test:
	go test ./pkg/... -timeout 1m30s

.PHONY: lint
lint:
	golangci-lint run ./cmd/... ./pkg/...

.PHONY: fmt
fmt:
	go fmt ./...

# The -tags builds static binary on linux.
# On macOS the binary is NOT static though.
# https://github.com/golang/go/issues/26492#issuecomment-635563222
.PHONY: build
build:
	go build -o authgear -tags 'osusergo netgo static_build' ./cmd/auth

.PHONY: check-tidy
check-tidy:
	$(MAKE) generate; go mod tidy; git status --porcelain | grep '.*'; test $$? -eq 1

.PHONY: build-image
build-image:
	docker build -f Dockerfile . -t authgear

# Compile mjml and print to stdout.
# You should capture the output and update the default template in Go code.
# For example,
# make html-email NAME=./templates/forgot_password_email.mjml | pbcopy
.PHONY: html-email
html-email:
	./scripts/npm/node_modules/.bin/mjml -l strict $(NAME)
