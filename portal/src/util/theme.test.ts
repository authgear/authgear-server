/* global describe, it, expect */
import { parse } from "postcss";
import { getShades, getTheme } from "./theme";

describe("getShades", () => {
  it("gives the same result as https://fabricweb.z5.web.core.windows.net/pr-deploy-site/refs/heads/master/theming-designer/index.html does", () => {
    const expected = [
      "#0078d4",
      "#f3f9fd",
      "#d0e7f8",
      "#a9d3f2",
      "#5ca9e5",
      "#1a86d9",
      "#006cbe",
      "#005ba1",
      "#004377",
    ];
    const actual = getShades("#0078d4");
    expect(actual).toEqual(expected);
  });
});

const CSS = `
:root {
  --color-primary-unshaded: #176df3;
  --color-primary-shaded-1: #f5f9fe;
  --color-primary-shaded-2: #d8e6fd;
  --color-primary-shaded-3: #b7d1fb;
  --color-primary-shaded-4: #70a4f7;
  --color-primary-shaded-5: #317bf4;
  --color-primary-shaded-6: #1460da;
  --color-primary-shaded-7: #1151b8;
  --color-primary-shaded-8: #0c3c88;

  --color-text-unshaded: #000000;
  --color-text-shaded-1: #898989;
  --color-text-shaded-2: #737373;
  --color-text-shaded-3: #595959;
  --color-text-shaded-4: #373737;
  --color-text-shaded-5: #2f2f2f;
  --color-text-shaded-6: #252525;
  --color-text-shaded-7: #151515;
  --color-text-shaded-8: #0b0b0b;

  --color-background-unshaded: #ffffff;
  --color-background-shaded-1: #767676;
  --color-background-shaded-2: #a6a6a6;
  --color-background-shaded-3: #c8c8c8;
  --color-background-shaded-4: #d0d0d0;
  --color-background-shaded-5: #dadada;
  --color-background-shaded-6: #eaeaea;
  --color-background-shaded-7: #f4f4f4;
  --color-background-shaded-8: #f8f8f8;
}
@media (prefers-color-scheme: dark) {
  :root {
    --color-primary-unshaded: #317BF4;
    --color-primary-shaded-1: #f6faff;
    --color-primary-shaded-2: #dde9fd;
    --color-primary-shaded-3: #bfd7fc;
    --color-primary-shaded-4: #81aff9;
    --color-primary-shaded-5: #498bf6;
    --color-primary-shaded-6: #2c70dc;
    --color-primary-shaded-7: #255eba;
    --color-primary-shaded-8: #1b4589;

    --color-text-unshaded: #ffffff;
    --color-text-shaded-1: #767676;
    --color-text-shaded-2: #a6a6a6;
    --color-text-shaded-3: #c8c8c8;
    --color-text-shaded-4: #d0d0d0;
    --color-text-shaded-5: #dadada;
    --color-text-shaded-6: #eaeaea;
    --color-text-shaded-7: #f4f4f4;
    --color-text-shaded-8: #f8f8f8;

    --color-background-unshaded: #000000;
    --color-background-shaded-1: #898989;
    --color-background-shaded-2: #737373;
    --color-background-shaded-3: #595959;
    --color-background-shaded-4: #373737;
    --color-background-shaded-5: #2f2f2f;
    --color-background-shaded-6: #252525;
    --color-background-shaded-7: #151515;
    --color-background-shaded-8: #0b0b0b;
  }
}
`;

describe("getTheme", () => {
  it("extracts theme color", () => {
    const root = parse(CSS);
    const actual = getTheme(root.nodes);
    expect(actual).toEqual({
      lightModePrimaryColor: "#176df3",
      lightModeTextColor: "#000000",
      lightModeBackgroundColor: "#ffffff",
      darkModePrimaryColor: "#317BF4",
      darkModeTextColor: "#ffffff",
      darkModeBackgroundColor: "#000000",
    });
  });

  it("returns null", () => {
    const actual = getTheme([]);
    expect(actual).toEqual(null);
  });
});
