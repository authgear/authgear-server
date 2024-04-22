from fontTools.subset import Subsetter, parse_unicodes, load_font, save_font

def subset_font(
  input_file_object_or_path,
  output_file_object_or_path,
  flavor,
  unicode_range,
  layout_closure=True,
):
  s = Subsetter()

  # Suppress warning due to missing tables
  s.options.drop_tables += ["FFTM"]

  # Disable layout closure to avoid including unwanted glyphs
  s.options.layout_closure = layout_closure

  if flavor != None and flavor != 'ttf':
    s.options.flavor = flavor

  font = load_font(input_file_object_or_path, s.options, dontLoadGlyphNames=False)
  s.populate(unicodes=parse_unicodes(unicode_range))
  s.subset(font)

  save_font(font, output_file_object_or_path, s.options)
