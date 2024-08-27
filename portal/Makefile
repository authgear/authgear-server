.PHONY: check-tidy
check-tidy:
	git status --porcelain | grep '.*'; test $$? -eq 1

.PHONY: fmt
fmt:
	npm run eslint:format
	npm run stylelint:format
	npm run prettier:format
	npm run gentype
	npm run generateCodesplit

.PHONY: lint
lint:
	npm run typecheck
	npm run eslint
	npm run stylelint
	npm run prettier
