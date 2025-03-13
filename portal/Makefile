.PHONY: check-tidy
check-tidy:
	git status --porcelain | grep '.*'; test $$? -eq 1

.PHONY: fmt
fmt:
	npm run eslint:format
	npm run stylelint:format
	npm run prettier:format
	npm run gentype

.PHONY: lint
lint:
	npm run typecheck
	npm run eslint
	npm run stylelint
	npm run prettier

.PHONY: outdated
outdated:
	@# Duplicate the output of npm outdated to stderr so that we can see the details in the build log.
	@# Also suppress the stdout of jq because it just prints either true or false,
	@# what we cares is just the exit code given by -e.
	npm outdated browserslist caniuse-lite stylelint-no-unsupported-browser-features --json | tee /dev/stderr | jq -e '. == {}' >/dev/null

.PHONY: deploy-storybook
deploy-storybook: export CHROMATIC_PROJECT_TOKEN := $(CHROMATIC_PROJECT_TOKEN)
deploy-storybook:
	npx chromatic --exit-zero-on-changes
