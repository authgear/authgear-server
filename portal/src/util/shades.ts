// A dependency-free port of FluentUI's color shading algorithm
// (@fluentui/react utilities/color, MIT licensed), kept so the generated
// AuthUI brand shades stay byte-identical to what the portal produced when it
// used FluentUI's ThemeGenerator. Verified against golden fixtures generated
// from the original implementation (see theme.test.ts).
//
// Behavioral narrowing vs the original getColorFromString: the original fell
// back to a DOM-based getComputedStyle trick for named colors ("red"); this
// port parses only #rgb, #rrggbb, rgb(a)(), and hsl(a)() strings and returns
// null otherwise. The portal only ever feeds it color-picker output.

const MAX_COLOR_RGB = 255;
const MAX_COLOR_ALPHA = 100;
const MAX_COLOR_SATURATION = 100;
const MAX_COLOR_VALUE = 100;

export interface RGBA {
  r: number;
  g: number;
  b: number;
  a: number;
}

interface HSV {
  h: number;
  s: number;
  v: number;
}

/** Shades of a given color, from softest to strongest. */
export enum Shade {
  Unshaded = 0,
  Shade1 = 1,
  Shade2 = 2,
  Shade3 = 3,
  Shade4 = 4,
  Shade5 = 5,
  Shade6 = 6,
  Shade7 = 7,
  Shade8 = 8,
}

// Luminance multiplier tables from FluentUI's shades.ts.
const WhiteShadeTable = [
  0.537, 0.349, 0.216, 0.184, 0.145, 0.082, 0.043, 0.027,
]; // white fg
const BlackTintTable = [0.537, 0.45, 0.349, 0.216, 0.184, 0.145, 0.082, 0.043]; // black fg
const LumTintTable = [0.88, 0.77, 0.66, 0.55, 0.44, 0.33, 0.22, 0.11]; // light (strongen all)
const LumShadeTable = [0.11, 0.22, 0.33, 0.44, 0.55, 0.66, 0.77, 0.88]; // dark (soften all)
const ColorTintTable = [0.96, 0.84, 0.7, 0.4, 0.12]; // default soften
const ColorShadeTable = [0.1, 0.24, 0.44]; // default strongen
const LowLuminanceThreshold = 0.2;
const HighLuminanceThreshold = 0.8;

function clamp(value: number, max: number, min: number = 0): number {
  return value < min ? min : value > max ? max : value;
}

function rgb2hsv(r: number, g: number, b: number): HSV {
  let h: number;
  const max = Math.max(r, g, b);
  const min = Math.min(r, g, b);
  const delta = max - min;

  if (delta === 0) {
    h = 0;
  } else if (r === max) {
    h = ((g - b) / delta) % 6;
  } else if (g === max) {
    h = (b - r) / delta + 2;
  } else {
    h = (r - g) / delta + 4;
  }

  h = Math.round(h * 60);
  if (h < 0) {
    h += 360;
  }

  const s = Math.round((max === 0 ? 0 : delta / max) * 100);
  const v = Math.round((max / MAX_COLOR_RGB) * 100);
  return { h, s, v };
}

function hsv2hsl(
  h: number,
  s: number,
  v: number
): { h: number; s: number; l: number } {
  s /= MAX_COLOR_SATURATION;
  v /= MAX_COLOR_VALUE;

  let l = (2 - s) * v;
  let sl = s * v;
  sl /= l <= 1 ? l : 2 - l;
  sl = sl || 0;
  l /= 2;
  return { h, s: sl * 100, l: l * 100 };
}

function hsl2hsv(h: number, s: number, l: number): HSV {
  s *= (l < 50 ? l : 100 - l) / 100;
  const v = l + s;
  return {
    h,
    s: v === 0 ? 0 : ((2 * s) / v) * 100,
    v,
  };
}

function hsv2rgb(
  h: number,
  s: number,
  v: number
): { r: number; g: number; b: number } {
  s = s / 100;
  v = v / 100;

  let rgb: number[] = [];
  const c = v * s;
  const hh = h / 60;
  const x = c * (1 - Math.abs((hh % 2) - 1));
  const m = v - c;

  switch (Math.floor(hh)) {
    case 0:
      rgb = [c, x, 0];
      break;
    case 1:
      rgb = [x, c, 0];
      break;
    case 2:
      rgb = [0, c, x];
      break;
    case 3:
      rgb = [0, x, c];
      break;
    case 4:
      rgb = [x, 0, c];
      break;
    case 5:
      rgb = [c, 0, x];
      break;
  }

  return {
    r: Math.round(MAX_COLOR_RGB * (rgb[0] + m)),
    g: Math.round(MAX_COLOR_RGB * (rgb[1] + m)),
    b: Math.round(MAX_COLOR_RGB * (rgb[2] + m)),
  };
}

function hsl2rgb(
  h: number,
  s: number,
  l: number
): { r: number; g: number; b: number } {
  const hsv = hsl2hsv(h, s, l);
  return hsv2rgb(hsv.h, hsv.s, hsv.v);
}

function rgbToPaddedHex(num: number): string {
  num = clamp(num, MAX_COLOR_RGB);
  const hex = num.toString(16);
  return hex.length === 1 ? "0" + hex : hex;
}

function rgb2hex(r: number, g: number, b: number): string {
  return [rgbToPaddedHex(r), rgbToPaddedHex(g), rgbToPaddedHex(b)].join("");
}

/**
 * A CSS color string from components: `rgba()` when alpha < 100, otherwise
 * `#rrggbb`. Matches FluentUI's `_rgbaOrHexString`.
 */
function rgbaOrHexString(r: number, g: number, b: number, a: number): string {
  return a === MAX_COLOR_ALPHA
    ? `#${rgb2hex(r, g, b)}`
    : `rgba(${r}, ${g}, ${b}, ${a / MAX_COLOR_ALPHA})`;
}

function parseRGBA(str: string): RGBA | null {
  const match = /^rgb(a?)\(([\d., ]+)\)$/.exec(str);
  if (match) {
    const hasAlpha = !!match[1];
    const expectedPartCount = hasAlpha ? 4 : 3;
    const parts = match[2].split(/ *, */).map(Number);
    if (parts.length === expectedPartCount) {
      return {
        r: parts[0],
        g: parts[1],
        b: parts[2],
        a: hasAlpha ? parts[3] * 100 : MAX_COLOR_ALPHA,
      };
    }
  }
  return null;
}

function parseHSLA(str: string): RGBA | null {
  const match = /^hsl(a?)\(([\d., ]+)\)$/.exec(str);
  if (match) {
    const hasAlpha = !!match[1];
    const expectedPartCount = hasAlpha ? 4 : 3;
    const parts = match[2].split(/ *, */).map(Number);
    if (parts.length === expectedPartCount) {
      const rgb = hsl2rgb(parts[0], parts[1], parts[2]);
      return {
        ...rgb,
        a: hasAlpha ? parts[3] * 100 : MAX_COLOR_ALPHA,
      };
    }
  }
  return null;
}

function parseHex6(str: string): RGBA | null {
  if (str[0] === "#" && str.length === 7 && /^#[\da-fA-F]{6}$/.test(str)) {
    return {
      r: parseInt(str.slice(1, 3), 16),
      g: parseInt(str.slice(3, 5), 16),
      b: parseInt(str.slice(5, 7), 16),
      a: MAX_COLOR_ALPHA,
    };
  }
  return null;
}

function parseHex3(str: string): RGBA | null {
  if (str[0] === "#" && str.length === 4 && /^#[\da-fA-F]{3}$/.test(str)) {
    return {
      r: parseInt(str[1] + str[1], 16),
      g: parseInt(str[2] + str[2], 16),
      b: parseInt(str[3] + str[3], 16),
      a: MAX_COLOR_ALPHA,
    };
  }
  return null;
}

/**
 * Parses a CSS color string (#rgb, #rrggbb, rgb(a)(), hsl(a)()) into RGBA
 * with alpha in [0, 100]. Returns null for anything else.
 */
export function parseCSSColor(str: string): RGBA | null {
  if (!str) {
    return null;
  }
  return parseRGBA(str) ?? parseHex6(str) ?? parseHex3(str) ?? parseHSLA(str);
}

function isWhite(c: RGBA): boolean {
  return (
    c.r === MAX_COLOR_RGB && c.g === MAX_COLOR_RGB && c.b === MAX_COLOR_RGB
  );
}

function isBlack(c: RGBA): boolean {
  return c.r === 0 && c.g === 0 && c.b === 0;
}

function darken(hsv: HSV, factor: number): HSV {
  return {
    h: hsv.h,
    s: hsv.s,
    v: clamp(hsv.v - hsv.v * factor, 100, 0),
  };
}

function lighten(hsv: HSV, factor: number): HSV {
  return {
    h: hsv.h,
    s: clamp(hsv.s - hsv.s * factor, 100, 0),
    v: clamp(hsv.v + (100 - hsv.v) * factor, 100, 0),
  };
}

/**
 * Given a color and a shade specification, generates the requested shade of
 * the color as a CSS color string, preserving the input's alpha. Port of
 * FluentUI's getShade with isInverted always false.
 */
export function getShadeString(color: RGBA, shade: Shade): string {
  if (shade === Shade.Unshaded) {
    return rgbaOrHexString(color.r, color.g, color.b, color.a);
  }

  const { h, s, v } = rgb2hsv(color.r, color.g, color.b);
  const hsl = hsv2hsl(h, s, v);
  let hsv: HSV = { h, s, v };
  const tableIndex = shade - 1;

  if (isWhite(color)) {
    hsv = darken(hsv, WhiteShadeTable[tableIndex]);
  } else if (isBlack(color)) {
    hsv = lighten(hsv, BlackTintTable[tableIndex]);
  } else if (hsl.l / 100 > HighLuminanceThreshold) {
    hsv = darken(hsv, LumShadeTable[tableIndex]);
  } else if (hsl.l / 100 < LowLuminanceThreshold) {
    hsv = lighten(hsv, LumTintTable[tableIndex]);
  } else {
    if (tableIndex < ColorTintTable.length) {
      hsv = lighten(hsv, ColorTintTable[tableIndex]);
    } else {
      hsv = darken(hsv, ColorShadeTable[tableIndex - ColorTintTable.length]);
    }
  }

  const rgb = hsv2rgb(hsv.h, hsv.s, hsv.v);
  return rgbaOrHexString(rgb.r, rgb.g, rgb.b, color.a);
}
