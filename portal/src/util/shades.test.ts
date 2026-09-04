import { describe, it, expect } from "@jest/globals";
import { parseCSSColor, rgbaOrHexString } from "./shades";

function normalize(value: string): string | undefined {
  const rgba = parseCSSColor(value);
  if (rgba == null) {
    return undefined;
  }
  return rgbaOrHexString(rgba.r, rgba.g, rgba.b, rgba.a);
}

describe("parseCSSColor", () => {
  it("accepts every format the FluentUI colour picker accepted", () => {
    expect(normalize("#176df3")).toEqual("#176df3");
    expect(normalize("#FFF")).toEqual("#ffffff");
    expect(normalize("rgb(23, 109, 243)")).toEqual("#176df3");
    expect(normalize("rgba(23, 109, 243, 0.5)")).toEqual(
      "rgba(23, 109, 243, 0.5)"
    );
  });

  // s/l are percentages in every spec-legal hsl() string. Before the `%` was
  // accepted, parseHSLA could not match any real hsl() value at all.
  it("accepts hsl() with percentage saturation and lightness", () => {
    expect(normalize("hsl(217, 89%, 52%)")).toEqual("#186bf2");
    expect(normalize("hsla(217, 89%, 52%, 0.5)")).toEqual(
      "rgba(24, 107, 242, 0.5)"
    );
  });

  it("rejects values it cannot represent", () => {
    expect(parseCSSColor("")).toBeNull();
    expect(parseCSSColor("#12")).toBeNull();
    expect(parseCSSColor("#1234567")).toBeNull();
    expect(parseCSSColor("not-a-color")).toBeNull();
  });
});
