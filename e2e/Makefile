.PHONY: build
build:
	go build -o $(BIN_NAME) ./cmd/$(TARGET)

.PHONY: run
run:
	./run.sh

.PHONY: setup
setup:
	./run.sh setup

.PHONY: tests
tests:
	./run.sh tests

.PHONY: teardown
teardown:
	./run.sh teardown
