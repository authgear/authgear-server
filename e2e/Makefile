include ../makefiles/go-mod-outdated.mk
include ../makefiles/govulncheck.mk

.PHONY: ci
ci:
	@# Check if any `focus: true` is present in the test files
	@# This is to prevent accidentally running only a subset of tests
	@! grep -r -l -E 'focus\s*:\strue' tests || (echo "Remove 'focus: true' from test files above." && exit 1)
	$(MAKE) run

.PHONY: clean
clean:
	rm -rf ./docker-compose-build-cache

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

# You need to turn on the containerd image store.
# See https://docs.docker.com/desktop/features/containerd/#enable-the-containerd-image-store
.PHONY: docker-compose-build
docker-compose-build:
	docker buildx bake --allow=fs.read=../postgres/postgres12 --file docker-compose.yaml --file docker-compose-build-cache.json
