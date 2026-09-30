// Design tokens for code that paints outside CSS (canvas charts).
//
// Radix defines its tokens on the .radix-themes element rather than on
// <html>, and on wide-gamut displays their values are oklch()/display-p3,
// which Chart.js cannot parse. Read from the theme root and normalize the
// value through a canvas so callers get a plain hex/rgba string. Non-color
// tokens (e.g. font families) are returned as-is.

function themeRoot(): Element {
  return document.querySelector(".radix-themes") ?? document.documentElement;
}

let scratch: CanvasRenderingContext2D | null | undefined;

function normalizeColor(value: string): string | null {
  if (scratch === undefined) {
    scratch = document.createElement("canvas").getContext("2d");
  }
  if (scratch == null) {
    return null;
  }
  // Setting an unparseable color leaves fillStyle unchanged.
  const sentinel = "#010203";
  scratch.fillStyle = sentinel;
  scratch.fillStyle = value;
  const normalized = scratch.fillStyle;
  return typeof normalized === "string" && normalized !== sentinel
    ? normalized
    : null;
}

export function readThemeToken(name: string, fallback: string): string {
  if (typeof window === "undefined") {
    return fallback;
  }
  const raw = getComputedStyle(themeRoot()).getPropertyValue(name).trim();
  if (raw === "") {
    return fallback;
  }
  return normalizeColor(raw) ?? raw;
}

export interface ChartThemeColors {
  accent: string;
  accentContrast: string;
  placeholder: string;
  neutralStrong: string;
  text: string;
  grid: string;
  surface: string;
}

// Fallbacks are the light-theme values, used only when no theme is mounted.
export function readChartThemeColors(): ChartThemeColors {
  return {
    accent: readThemeToken("--accent-9", "#176df3"),
    accentContrast: readThemeToken("--accent-contrast", "#ffffff"),
    placeholder: readThemeToken("--gray-5", "#e0e1e6"),
    neutralStrong: readThemeToken("--gray-12", "#1e1f24"),
    text: readThemeToken("--gray-11", "#62636c"),
    grid: readThemeToken("--gray-a4", "#000b3618"),
    surface: readThemeToken("--color-panel-solid", "#ffffff"),
  };
}
