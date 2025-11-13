DISPOSABLE_EMAIL_DOMAINS_URL ::= https://raw.githubusercontent.com/disposable-email-domains/disposable-email-domains/refs/heads/main/disposable_email_blocklist.conf
DISPOSABLE_EMAIL_DOMAINS_FILE ::= ./resources/authgear/disposable_email_domain_list.txt
FREE_EMAIL_DOMAINS_JSON_URL ::= https://raw.githubusercontent.com/Kikobeats/free-email-domains/refs/heads/master/domains.json
FREE_EMAIL_DOMAINS_FILE ::= ./resources/authgear/free_email_provider_domain_list.txt

CHECK_FILES_SCRIPT ::= $(abspath $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))/../scripts/sh/check-files-outdated.sh)

.PHONY: check-disposable-email-domains
check-disposable-email-domains:
	@$(CHECK_FILES_SCRIPT) compare-file "$(DISPOSABLE_EMAIL_DOMAINS_URL)" "$(DISPOSABLE_EMAIL_DOMAINS_FILE)"

.PHONY: check-free-email-domains
check-free-email-domains:
	@$(CHECK_FILES_SCRIPT) compare-json "$(FREE_EMAIL_DOMAINS_JSON_URL)" "$(FREE_EMAIL_DOMAINS_FILE)"

.PHONY: check-email-domains-update
check-email-domains-update: check-disposable-email-domains check-free-email-domains

.PHONY: update-disposable-email-domains
update-disposable-email-domains:
	@$(CHECK_FILES_SCRIPT) compare-file "$(DISPOSABLE_EMAIL_DOMAINS_URL)" "$(DISPOSABLE_EMAIL_DOMAINS_FILE)" --update

.PHONY: update-free-email-domains
update-free-email-domains:
	@$(CHECK_FILES_SCRIPT) compare-json "$(FREE_EMAIL_DOMAINS_JSON_URL)" "$(FREE_EMAIL_DOMAINS_FILE)" --update

.PHONY: update-email-domains
update-email-domains: update-disposable-email-domains update-free-email-domains
