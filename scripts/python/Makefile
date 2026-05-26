PIP := ./venv/bin/pip
PYTHON := ./venv/bin/python

# We use the python found in PATH to create a venv
# After that, we always use $(PIP) or $(PYTHON)
.PHONY: venv
venv:
	python -m venv venv
	$(PIP) install --upgrade pip
	$(PIP) install -r requirements.txt

.PHONY: generate-material-icons
generate-material-icons: venv
	$(PYTHON) ./subset_fonts/subset_material_icons.py

.PHONY: generate-twemoji-icons
generate-twemoji-icons: venv
	$(PYTHON) ./subset_fonts/subset_twemoji_icons.py

# `make generate-translations` uses the claude CLI with your Claude.ai subscription quota.
# Make sure you are logged in first: claude login
.PHONY: generate-translations
generate-translations: venv
	$(PYTHON) ./generate_translations.py
	@echo "Run 'make html-email' in project root to generate email templates."
