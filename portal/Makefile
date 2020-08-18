.PHONY: check-tidy
check-tidy:
	git status --porcelain | grep '.*'; test $$? -eq 1
