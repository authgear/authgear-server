.PHONY: go-mod-download
go-mod-download:
	@cp ./go.mod.tpl ./go.mod
	@go mod download

.PHONY: go-mod-outdated
go-mod-outdated: go-mod-download
	@# https://stackoverflow.com/questions/55866604/whats-the-go-mod-equivalent-of-npm-outdated
	@# Since go 1.21, this command will exit 2 when one of the dependencies require a go version newer than us.
	@# This implies we have to use the latest verion of Go whenever possible.
	@go list -u -m -f '{{if .Update}}{{if not .Indirect}}{{.}}{{end}}{{end}}' all

.PHONY: check-if-new-version-available
check-if-new-version-available:
	@# If grep matches something, it exits 0, otherwise it exits 1.
	@# In our case, we want to invert the exit code.
	$(MAKE) go-mod-outdated | grep "."; status_code=$$?; if [ $$status_code -eq 0 ]; then exit 1; else exit 0; fi;
