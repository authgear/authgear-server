.PHONY: vendor
vendor:
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.22.2
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

.PHONY: build
build:
	go build -o authgear ./cmd/auth

.PHONY: check-tidy
check-tidy:
	$(MAKE) generate; go mod tidy; git status --porcelain | grep '.*'; test $$? -eq 1

.PHONY: build-image
build-image:
	docker build -f Dockerfile . -t authgear
