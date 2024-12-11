.PHONY: check-if-new-version-available
check-if-new-version-available:
	@cp ./go.mod.tpl ./go.mod
	@go mod download
	@# If grep matches something, it exits 0, otherwise it exits 1.
	@# In our case, we want to invert the exit code.
	go list -u -m -f '{{if .Update}}{{if not .Indirect}}{{.}}{{end}}{{end}}' all | grep "."; status_code=$$?; if [ $$status_code -eq 0 ]; then exit 1; else exit 0; fi;
