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

# `make generate-translations` expects the environment variable ANTHROPIC_API_KEY to be set.
# You can set it at the CLI with
# ANTHROPIC_API_KEY=your_key make generate-translations
# Or create a .env file next to the Makefile. The .env file will be loaded dotenv in the script.
.PHONY: generate-translations
generate-translations: venv
	$(PYTHON) ./generate_translations.py
	@echo "Run 'make html-email' in project root to generate email templates."
