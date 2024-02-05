import os
import re
import requests
from bs4 import BeautifulSoup

from lib import subset_font

def parse_html_file(file_path):
  with open(file_path, 'r') as file:
    soup = BeautifulSoup(file.read(), 'html.parser')
    icons = soup.find_all(class_='material-icons')
    icon_names = []
    for icon in icons:
      if not icon.string:
        continue

      icon_text = icon.string.replace('\n', '').strip()
      if not icon_text:
        continue

      # Handle golang template syntax
      # e.g. {{ if .IsEmail }}email{{ else }}phone{{ end }} -> ['email', 'phone']
      if icon_text.startswith('{{'):
        matches = re.findall('}}([^{]*){{', icon_text)
        for match in matches:
            icon_names.append(match.strip())
      else:
          icon_names.append(icon_text)

    return icon_names

# Find all <i class="material-icons">icon_name</i> in authflowv2 templates
def parse_html_directory(directory_path):
  icon_names = []
  for root, dirs, files in os.walk(directory_path):
    for file in files:
      if file.endswith('.html'):
        file_path = os.path.join(root, file)
        icon_names.extend(parse_html_file(file_path))
    for dir in dirs:
      icon_names.extend(parse_html_directory(os.path.join(root, dir)))
  return set(icon_names)


# Find all `content: "icon_name"` in authflowv2 css files
def parse_css_directory(directory_path):
  icon_names = []
  for root, dirs, files in os.walk(directory_path):
    for file in files:
      if file.endswith('.css'):
        file_path = os.path.join(root, file)
        with open(file_path, 'r') as file:
          content = file.read()
          matches = re.findall(r'content: "([^"]+)"', content)
          icon_names.extend(matches)
    for dir in dirs:
      icon_names.extend(parse_css_directory(os.path.join(root, dir)))

  return set(icon_names)


if __name__ == '__main__':
  # Fetch codepoints for subsetting by unicode
  response = requests.get('https://raw.githubusercontent.com/google/material-design-icons/master/variablefont/MaterialSymbolsOutlined%5BFILL%2CGRAD%2Copsz%2Cwght%5D.codepoints')
  lines = response.text.split('\n')
  codepoints = {line.split()[0]: line.split()[1] for line in lines if line}

  # Find icon names in authflowv2 templates and css files
  html_directory_path = '../../resources/authgear/templates/en/web/authflowv2'
  css_directory_path = '../../authui/src/authflowv2'
  icon_names = parse_html_directory(html_directory_path) | parse_css_directory(css_directory_path)

  print(f'Found {len(icon_names)} icon names:')
  print(', '.join(icon_names))

  icon_codepoints = [codepoints[name] for name in icon_names if name in codepoints]

  _latin_unicode_range = 'U+0030-0039,U+0061-007A,U+005F,'  # 0-9, a-z, _
  unicode_range = _latin_unicode_range + ','.join([f'U+{codepoint.upper()}' for codepoint in icon_codepoints])

  # Subset ttf and woff2 fonts
  subset_font(
    '../../authui/src/authflowv2/icons/material-symbols-outlined.woff2',
    '../../authui/src/authflowv2/icons/material-symbols-outlined-subset.woff2',
    'woff2',
    unicode_range,
    layout_closure=False,
  )
  subset_font(
    '../../authui/src/authflowv2/icons/material-symbols-outlined.ttf',
    '../../authui/src/authflowv2/icons/material-symbols-outlined-subset.ttf',
    'ttf',
    unicode_range,
    layout_closure=False,
  )
