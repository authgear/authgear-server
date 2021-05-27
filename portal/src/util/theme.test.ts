/* global describe, it, expect */
import { parse, Root } from "postcss";
import {
  getShades,
  getLightTheme,
  getDarkTheme,
  getLightBannerConfiguration,
  getDarkBannerConfiguration,
  addLightTheme,
  addDarkTheme,
  addLightBannerConfiguration,
  addDarkBannerConfiguration,
  LightTheme,
  DarkTheme,
  DEFAULT_BANNER_CONFIGURATION,
} from "./theme";

const DEFAULT_LIGHT_THEME: LightTheme = {
  isDarkTheme: false,
  primaryColor: "#176df3",
  textColor: "#000000",
  backgroundColor: "#ffffff",
};

const DEFAULT_DARK_THEME: DarkTheme = {
  isDarkTheme: true,
  primaryColor: "#317BF4",
  textColor: "#ffffff",
  backgroundColor: "#000000",
};

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

.banner-frame {
  background-color: red;
  padding-top: 2px;
  padding-right: 3px;
  padding-bottom: 4px;
  padding-left: 5px;
}
.banner {
  width: initial;
  height: 1px;
}

@media (prefers-color-scheme: dark) {
  .banner-frame {
    background-color: blue;
    padding-top: 3px;
    padding-right: 4px;
    padding-bottom: 5px;
    padding-left: 6px;
  }
  .banner {
    width: initial;
    height: 2px;
  }
}
`;

describe("getLightTheme", () => {
  it("extracts theme color", () => {
    const root = parse(CSS);
    const actual = getLightTheme(root.nodes);
    expect(actual).toEqual({
      isDarkTheme: false,
      primaryColor: "#176df3",
      textColor: "#000000",
      backgroundColor: "#ffffff",
    });
  });

  it("returns null", () => {
    const actual = getDarkTheme([]);
    expect(actual).toEqual(null);
  });
});

describe("getDarkTheme", () => {
  it("extracts theme color", () => {
    const root = parse(CSS);
    const actual = getDarkTheme(root.nodes);
    expect(actual).toEqual({
      isDarkTheme: true,
      primaryColor: "#317BF4",
      textColor: "#ffffff",
      backgroundColor: "#000000",
    });
  });

  it("returns null", () => {
    const actual = getDarkTheme([]);
    expect(actual).toEqual(null);
  });
});

describe("getLightBannerConfiguration", () => {
  it("extracts banner configuration", () => {
    const root = parse(CSS);
    const actual = getLightBannerConfiguration(root.nodes);
    expect(actual).toEqual({
      width: "initial",
      height: "1px",
      paddingTop: "2px",
      paddingRight: "3px",
      paddingBottom: "4px",
      paddingLeft: "5px",
      backgroundColor: "red",
    });
  });

  it("returns null", () => {
    const actual = getLightBannerConfiguration([]);
    expect(actual).toEqual(null);
  });
});

describe("getDarkBannerConfiguration", () => {
  it("extracts banner configuration", () => {
    const root = parse(CSS);
    const actual = getDarkBannerConfiguration(root.nodes);
    expect(actual).toEqual({
      width: "initial",
      height: "2px",
      paddingTop: "3px",
      paddingRight: "4px",
      paddingBottom: "5px",
      paddingLeft: "6px",
      backgroundColor: "blue",
    });
  });

  it("returns null", () => {
    const actual = getDarkBannerConfiguration([]);
    expect(actual).toEqual(null);
  });
});

describe("addLightTheme", () => {
  it("renders theme into CSS", () => {
    const root = new Root();
    addLightTheme(root, DEFAULT_LIGHT_THEME);
    const actual = root.toResult().css;
    const expected = `:root {
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
    --color-background-shaded-8: #f8f8f8
}`;
    expect(actual).toEqual(expected);
  });
});

describe("addLightBannerConfiguration", () => {
  it("renders banner configuration into CSS", () => {
    const root = new Root();
    addLightBannerConfiguration(root, DEFAULT_BANNER_CONFIGURATION);
    const actual = root.toResult().css;
    const expected = `.banner {
    width: initial;
    height: 55px
}
.banner-frame {
    padding-top: 16px;
    padding-right: 16px;
    padding-bottom: 16px;
    padding-left: 16px;
    background-color: transparent
}`;
    expect(actual).toEqual(expected);
  });
});

describe("addDarkBannerConfiguration", () => {
  it("renders banner configuration into CSS", () => {
    const root = new Root();
    addDarkBannerConfiguration(root, DEFAULT_BANNER_CONFIGURATION);
    const actual = root.toResult().css;
    const expected = `@media (prefers-color-scheme: dark) {
    .banner {
        width: initial;
        height: 55px
    }
    .banner-frame {
        padding-top: 16px;
        padding-right: 16px;
        padding-bottom: 16px;
        padding-left: 16px;
        background-color: transparent
    }
}`;
    expect(actual).toEqual(expected);
  });
});

describe("addDarkTheme", () => {
  it("renders theme into CSS", () => {
    const root = new Root();
    addDarkTheme(root, DEFAULT_DARK_THEME);
    const actual = root.toResult().css;
    const expected = `@media (prefers-color-scheme: dark) {
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
        --color-background-shaded-8: #0b0b0b
    }
}`;
    expect(actual).toEqual(expected);
  });
});
