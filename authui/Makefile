.PHONY: check-tidy
check-tidy:
	make fmt
	git status --porcelain | grep '.*'; test $$? -eq 1

.PHONY: fmt
fmt:
	npm run eslint:format
	npm run stylelint:format
	npm run prettier:format

.PHONY: lint
lint:
	npm run eslint
	npm run stylelint
	npm run prettier
	npm run typecheck
