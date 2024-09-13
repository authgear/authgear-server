from lib import subset_font


if __name__ == '__main__':
  unicode_range = 'U+1F1E6-1F1FF'

  # Subset ttf and woff2 fonts for Twemoji.Mozilla
  subset_font(
    '../../authui/src/authflowv2/icons/Twemoji.Mozilla.woff2',
    '../../authui/src/authflowv2/icons/Twemoji.Mozilla-subset.woff2',
    'woff2',
    unicode_range,
  )
  subset_font(
    '../../authui/src/authflowv2/icons/Twemoji.Mozilla.ttf',
    '../../authui/src/authflowv2/icons/Twemoji.Mozilla-subset.ttf',
    'ttf',
    unicode_range,
  )
