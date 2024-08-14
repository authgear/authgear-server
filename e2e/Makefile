.PHONY: ci
ci:
	@# Check if any `focus: true` is present in the test files
	@# This is to prevent accidentally running only a subset of tests
	@! grep -r -l -E 'focus\s*:\strue' tests || (echo "Remove 'focus: true' from test files above." && exit 1)
	make run

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
