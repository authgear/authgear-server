PYTHON_VENV = venv/bin

.PHONT: generate-material-icons
generate-material-icons:
	python -m venv venv
	${PYTHON_VENV}/pip install -q -r requirements.txt
	${PYTHON_VENV}/python ./subset_material_icons.py

.PHONY: generate-twemoji-icons
generate-twemoji-icons:
	python -m venv venv
	${PYTHON_VENV}/pip install -q -r requirements.txt
	${PYTHON_VENV}/python ./subset_twemoji_icons.py
