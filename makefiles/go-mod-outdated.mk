.PHONY: go-mod-outdated
go-mod-outdated: THIS_DIR ::= $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
go-mod-outdated:
	$(THIS_DIR)/../scripts/sh/go-mod-outdated.sh
