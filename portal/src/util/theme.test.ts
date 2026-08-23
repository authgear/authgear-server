import { deriveColors, getShades } from "./theme";
import golden from "./__fixtures__/fluent-shades-golden.json";

// The fixtures were generated with the original FluentUI ThemeGenerator
// implementation. The ported algorithm must reproduce them exactly so that
// the AuthUI brand shades generated for existing projects do not change.
describe("shades port stays identical to FluentUI ThemeGenerator", () => {
  const entries = Object.entries(golden) as [
    string,
    {
      deriveColors: { original: string; variant: string };
      getShades: string[];
    }
  ][];

  test.each(entries)("deriveColors(%s)", (color, expected) => {
    expect(deriveColors(color)).toEqual(expected.deriveColors);
  });

  test.each(entries)("getShades(%s)", (color, expected) => {
    expect(getShades(color)).toEqual(expected.getShades);
  });

  it("returns null for unparseable colors", () => {
    expect(deriveColors("not-a-color")).toBeNull();
    expect(deriveColors("")).toBeNull();
  });
});
