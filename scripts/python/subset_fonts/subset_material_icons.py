import glob
import re

from lib import subset_font

def filter_icon_names_from_path(icon_names, path):
  filtered_icon_names = set()
  regex = re.compile(rf'\W({"|".join(icon_names)})\W')

  for file in glob.iglob(path, recursive=True):
    with open(file, 'r') as f:
      content = f.read()
      matched_icon_names = set(re.findall(regex, content))
      filtered_icon_names |= matched_icon_names

  return filtered_icon_names

if __name__ == '__main__':
  # Get codepoints for subsetting by unicode
  codepoint_file = '../../authui/src/authflowv2/icons/material-symbols-outlined.codepoints'
  with open(codepoint_file, 'r') as file:
    codepoints_dict = {name: codepoint for name, codepoint in (line.split() for line in file)}
    icon_names = codepoints_dict.keys()

  # Find icon names in authflowv2 templates and css files
  html_directory_path_authflowv2 = '../../resources/authgear/templates/en/web/authflowv2/**/*.html'
  html_directory_path_settingsv2 = '../../resources/authgear/templates/en/web/settingsv2/**/*.html'
  css_directory_path = '../../authui/src/authflowv2/**/*.css'
  icon_name_list_path = './subset_fonts/material-icons.txt'

  filtered_icon_names = \
    filter_icon_names_from_path(icon_names, html_directory_path_authflowv2) | \
    filter_icon_names_from_path(icon_names, html_directory_path_settingsv2) | \
    filter_icon_names_from_path(icon_names, css_directory_path)

  filtered_codepoints = [codepoints_dict[name] for name in filtered_icon_names if name in codepoints_dict]

  _latin_unicode_range = 'U+0030-0039,U+0061-007A,U+005F,'  # 0-9, a-z, _
  unicode_range = _latin_unicode_range + ','.join([f'U+{codepoint.upper()}' for codepoint in filtered_codepoints])

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

  # Update icon name list
  with open(icon_name_list_path, 'w') as file:
    for icon_name in sorted(filtered_icon_names):
      print(f"{icon_name}", file=file)

  print(f'\nUpdated {icon_name_list_path} with latest list.')
