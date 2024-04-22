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
