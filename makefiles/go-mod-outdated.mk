.PHONY: go-mod-outdated
go-mod-outdated:
	# https://stackoverflow.com/questions/55866604/whats-the-go-mod-equivalent-of-npm-outdated
	# Since go 1.21, this command will exit 2 when one of the dependencies require a go version newer than us.
	# This implies we have to use the latest verion of Go whenever possible.
	go list -u -m -f '{{if .Update}}{{if not .Indirect}}{{.}}{{end}}{{end}}' all
