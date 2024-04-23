import collections
import concurrent.futures
import os
import re
import anthropic
import json
import json_repair
import regex
import logging
from dotenv import load_dotenv

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')

load_dotenv()

LOCALE_DICT = {
  "zh-HK": "Chinese (Hong Kong)",
  "zh-TW": "Chinese (Taiwan)",
  "zh-CN": "Chinese (China)",
  "ko": "Korean",
  "ja": "Japanese",
  "vi": "Vietnamese",
  "th": "Thai",
  "ms": "Malay",
  "fil": "Filipino (Tagalog)",
  "id": "Indonesian",

  # "en-GB": "English (UK)",
  # "en-US": "English (US)",
  "fr": "French",
  "es-ES": "Spanish (Spain)",
  "es": "Spanish (Latin America)",
  "it": "Italian",
  "pt-PT": "Portuguese (Portugal)",
  "pt-BR": "Portuguese (Brazil)",
  "de": "German",
  "pl": "Polish",
  "nl": "Dutch",
  "el": "Greek"
}

def find_missing_keys(default_translation, current_translation):
  def key_is_reserved(key):
    return key.startswith("language-")
  missing_keys = [key for key in default_translation.keys() if key not in current_translation and not key_is_reserved(key)]
  return collections.OrderedDict((key, default_translation[key]) for key in missing_keys)

def chunk_messages(messages: dict[str, str | dict[str, str]], chunk_size: int) -> list[dict[str, str | dict[str, str]]]:
  keys = list(messages.keys())
  return [collections.OrderedDict((k, messages[k]) for k in keys[i:i + chunk_size]) for i in range(0, len(keys), chunk_size)]

def auto_translate(messages: dict[str, str | dict[str, str]], locale, chunk_size):
  if len(messages.keys()) == 0:
    return {}

  chunks = chunk_messages(messages, chunk_size)
  translated_messages = collections.OrderedDict()

  for idx, chunk in enumerate(chunks):
    logging.info(f'{locale} | Starting translation of chunk {idx + 1}/{len(chunks)}')

    prompt = f"""
    - translate the values in the JSON from English into {locale}.
    - Don't translate the phrase "Passkey", keep it as is.
    - Return only the JSON and nothing else
    - Escape astrophes (') with double astrophes ('')
    """

    client = anthropic.Anthropic()
    message = client.messages.create(
      model="claude-3-haiku-20240307",
      max_tokens=4000,
      temperature=0,
      system=prompt,
      messages=[
        {
          "role": "user",
          "content": [
            {
              "type": "text",
              "text": json.dumps(chunk, indent=2)
            }
          ]
        }
      ]
    )
    result = message.content[0].text

    logging.info(f'{locale} | Translation result: {result}')

    json_match = regex.search(r'{(?:[^{}]|(?R))*}', result)
    if not json_match:
      raise ValueError('Failed to extract JSON from translation result.')

    json_str = json_match.group()
    translated_chunk = json_repair.loads(json_str)
    translated_messages.update(translated_chunk) # type: ignore

    logging.info(f'{locale} | Finished translation of chunk {idx + 1}/{len(chunks)}')

    yield translated_messages

def ensure_path_exists(path):
  directory = os.path.dirname(path)
  if not os.path.exists(directory):
    os.makedirs(directory)

  if not os.path.isfile(path):
    with open(path, 'w'):
        pass

def update_translation(locale: str, default_translation_file: str, locale_translation_file: str, chunk_size):
  logging.info(f'{locale} | Updating {locale_translation_file} with latest keys.')

  with open(default_translation_file, 'r') as file:
    default_translation = json.load(file, object_pairs_hook=collections.OrderedDict)

  # Create locale translation file if it doesn't exist
  ensure_path_exists(locale_translation_file)

  with open(locale_translation_file, 'r') as file:
    try:
      locale_translation = json.load(file, object_pairs_hook=collections.OrderedDict)
    except json.JSONDecodeError:
      locale_translation = collections.OrderedDict()

  missing_keys = find_missing_keys(default_translation, locale_translation)

  logging.info(f'{locale} | Found {len(missing_keys)} missing keys in {locale_translation_file}.')

  def fix_translation_json(translation):
    for k, v in translation.items():
      if isinstance(v, str):
        # Fix escaped backslashes
        translation[k] = re.sub(r'\\{2}', r'\\', translation[k])

    return translation

  def save_translation_file(default_translation, updated_translation, file):
    translation = fix_translation_json(updated_translation)
    translation = collections.OrderedDict((k, translation[k]) for k in default_translation.keys() if k in translation)
    with open(file, 'w') as file:
      json.dump(translation, file, indent=2, ensure_ascii=False)
      file.write('\n')

  for translated_messages in auto_translate(messages=missing_keys, locale=locale, chunk_size=chunk_size):
    locale_translation.update(translated_messages)
    save_translation_file(default_translation, locale_translation, locale_translation_file)

  # Insert default translation for reserved keys (e.g. language-*)
  upodated_translation = default_translation.copy()
  upodated_translation.update(locale_translation)
  save_translation_file(default_translation, upodated_translation, locale_translation_file)

  missing_keys = find_missing_keys(default_translation, locale_translation)
  if len(missing_keys) > 0:
    logging.error(f'{locale} | Failed to translate the following keys: {missing_keys.keys()}.')

  logging.info(f'{locale} | Updated {locale_translation_file} with latest keys.')


def make_update_locale_fn(chunk_size):
  def update_locale(locale):
    try:
      logging.info(f'{locale} | Starting translation for {locale} ({LOCALE_DICT[locale]})')

      # HTML template translation
      update_translation(
        locale=locale,
        default_translation_file=f'../../resources/authgear/templates/en/translation.json',
        locale_translation_file=f'../../resources/authgear/templates/{locale}/translation.json',
        chunk_size=chunk_size
      )

      # Email / SMS template translation
      update_translation(
        locale=locale,
        default_translation_file=f'../../resources/authgear/templates/en/messages/translation.json',
        locale_translation_file=f'../../resources/authgear/templates/{locale}/messages/translation.json',
        chunk_size=chunk_size
      )

      logging.info(f'{locale} | Finished translation for {locale} ({LOCALE_DICT[locale]})')
    except Exception as e:
      logging.error(f'{locale} | Failed to update translations for {locale}: {e}')

  return update_locale

if __name__ == '__main__':
  max_workers = 2
  chunk_size = 10
  locales = [locale for locale in LOCALE_DICT.keys()]

  with concurrent.futures.ThreadPoolExecutor(max_workers=max_workers) as executor:
    executor.map(make_update_locale_fn(chunk_size), locales)
