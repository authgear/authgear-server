.PHONY: run
run:
	./run.sh

# Following targets are partial steps of the run target for debugging

.PHONY: setup
setup:
	./run.sh setup

.PHONY: tests
tests:
	./run.sh tests

.PHONY: teardown
teardown:
	./run.sh teardown

.PHONY: dumpschema
dumpschema:
	go run ./cmd/dumpschema
