PYTHON_VENV = venv/bin
PATH := $(PYTHON_VENV):$(PATH)

.PHONY: venv
venv:
	python -m venv venv
	. venv/bin/activate
	which python
	pip install -r requirements.txt

.PHONY: generate-material-icons
generate-material-icons: venv
	python ./subset_fonts/subset_material_icons.py

.PHONY: generate-twemoji-icons
generate-twemoji-icons: venv
	python ./subset_fonts/subset_twemoji_icons.py

# `make generate-translations` expects the environment variable ANTHROPIC_API_KEY and CLDR_LOCALNAMES_PATH to be set.
# You can set it at the CLI with
# ANTHROPIC_API_KEY=your_key CLDR_LOCALNAMES_PATH=your_path make generate-translations
# Or create a .env file next to the Makefile. The .env file will be loaded dotenv in the script.
.PHONY: generate-translations
generate-translations: venv
	python ./generate_translations.py
	@echo "Run 'make html-email' in project root to generate email templates."
