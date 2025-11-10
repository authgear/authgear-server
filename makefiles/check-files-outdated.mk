DISPOSABLE_EMAIL_DOMAINS_URL ::= https://raw.githubusercontent.com/disposable-email-domains/disposable-email-domains/refs/heads/main/disposable_email_blocklist.conf
DISPOSABLE_EMAIL_DOMAINS_FILE ::= ./resources/authgear/disposable_email_domain_list.txt

CHECK_FILES_SCRIPT ::= $(abspath $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))/../scripts/sh/check-files-outdated.sh)

.PHONY: check-disposable-email-domains
check-disposable-email-domains:
	@$(CHECK_FILES_SCRIPT) compare-file "$(DISPOSABLE_EMAIL_DOMAINS_URL)" "$(DISPOSABLE_EMAIL_DOMAINS_FILE)"

.PHONY: check-email-domains-update
check-email-domains-update: check-disposable-email-domains

.PHONY: update-disposable-email-domains
update-disposable-email-domains:
	@$(CHECK_FILES_SCRIPT) compare-file "$(DISPOSABLE_EMAIL_DOMAINS_URL)" "$(DISPOSABLE_EMAIL_DOMAINS_FILE)" --update

.PHONY: update-email-domains
update-email-domains: update-disposable-email-domains
