import json
import subprocess
from typing import Dict


def replace_old_key_with_new_key(key_map: Dict[str, str]):
    def construct_html_bash_script(old_key: str, new_key: str) -> str:
        new_key.replace(r".", r"\.")
        return f'find resources/authgear/templates/en/web/authflowv2 -type f -name "*.html" -print0 | xargs -0 gsed -i "s/\\b{old_key}\\b/{new_key}/g"'

    def construct_json_bash_script(old_key: str, new_key: str) -> str:
        new_key.replace(r".", r"\.")
        return f'find resources/authgear/templates/ -type f -name "translation.json" -print0 | xargs -0 gsed -i "s/\\b{old_key}\\b/{new_key}/g"'

    for old_key, new_key in key_map.items():
        html_bash_script = construct_html_bash_script(old_key, new_key)
        json_bash_script = construct_json_bash_script(old_key, new_key)
        subprocess.check_output(html_bash_script, shell=True)
        subprocess.check_output(json_bash_script, shell=True)


def add_new_key_while_keeping_old_key(custom_key_map: Dict[str, str]):
    def construct_html_bash_script(old_key: str, new_key: str) -> str:
        return rf'find resources/authgear/templates/en/web/authflowv2 -type f -name "*.html" -print0 | xargs -0 gsed -i "s/\b{old_key}\b/{new_key}/g"'

    def construct_json_bash_script(old_key: str, new_key: str) -> str:
        return rf'find resources/authgear/templates/ -type f -name "translation.json" -print0 | xargs -0 gsed -E -i "s/\"{old_key}\": \"(.+)\",/\"{old_key}\": \"\1\",\n  \"{new_key}\": \"\1\",/"'

    for old_key, new_key in custom_key_map.items():
        json_bash_script = construct_json_bash_script(old_key, new_key)
        html_bash_script = construct_html_bash_script(old_key, new_key)
        subprocess.check_output(json_bash_script, shell=True)
        subprocess.check_output(html_bash_script, shell=True)


def convert():
    key_map = dict()
    with open(
        "translation-key-map.json",
    ) as f:
        key_map: Dict[str, str] = json.load(f)
        add_key_map: Dict[str, str] = dict()
        replace_key_map: Dict[str, str] = dict()

        for old_key, new_key in key_map.items():
            if old_key.startswith("v2-"):
                replace_key_map[old_key] = new_key
            else:
                add_key_map[old_key] = new_key

        replace_old_key_with_new_key(replace_key_map)
        add_new_key_while_keeping_old_key(add_key_map)


if __name__ == "__main__":
    convert()
