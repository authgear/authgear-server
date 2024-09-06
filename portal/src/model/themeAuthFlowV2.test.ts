import { describe, it, expect } from "@jest/globals";

import { parse as parseCSS } from "postcss";
import {
  ColorStyleProperty,
  CssAstVisitor,
  CustomisableThemeStyleGroup,
  DEFAULT_LIGHT_THEME,
  StyleCssVisitor,
  StyleGroup,
  ThemeTargetSelector,
} from "./themeAuthFlowV2";

describe("StyleCssVisitor", () => {
  it("should parse stylesheet into styles", () => {
    const styleGroup = new StyleGroup({
      button: new StyleGroup({
        textColor: new ColorStyleProperty("-—primary-btn__bg-color", "red"),
      }),
    });

    const cssStyleSheet = `
      :root {
        -—primary-btn__bg-color: blue;
      }

      .other-selector {
        flex: 1;
      }
    `;

    const styleVisitor = new StyleCssVisitor(":root", styleGroup);
    const styles = styleVisitor.getStyle(parseCSS(cssStyleSheet));

    expect(styles).toEqual({
      button: {
        textColor: "blue",
      },
    });
  });

  it("should parse authui stylesheet", () => {
    const authuiStyleSheet = `:root {
    --layout__bg-color: #ffffff;
    --alignment-card: center;
    --layout-padding-left: 0;
    --layout-padding-right: 0,;
    --primary-btn__bg-color: #176df3;
    --primary-btn__bg-color--active: #1151b8;
    --primary-btn__bg-color--hover: #1151b8;
    --primary-btn__text-color: #ffffff;
    --primary-btn__border-radius: 0.875em;
    --secondary-btn__border-radius: 0.875em;
    --input__border-radius: 0.875em;
    --phone-input__trigger-border-radius: 0.875em;
    --icon-color: #176df3;
    --color-link: #176df3;
    --color-link--active: #1151b8;
    --color-link--hover: #1151b8;
    --body-text__link-text-decoration: underline;
    --brand-logo__height: 2.5rem
}`;
    const customisableThemeStyleGroup = new CustomisableThemeStyleGroup(
      DEFAULT_LIGHT_THEME
    );
    const styleVisitor = new StyleCssVisitor(
      ThemeTargetSelector.Light,
      customisableThemeStyleGroup
    );
    const styles = styleVisitor.getStyle(parseCSS(authuiStyleSheet));
    expect(styles).toEqual({
      page: {
        backgroundColor: "#ffffff",
      },
      card: {
        alignment: "center",
      },
      primaryButton: {
        backgroundColor: "#176df3",
        backgroundColorActive: "#1151b8",
        backgroundColorHover: "#1151b8",
        labelColor: "#ffffff",
        borderRadius: {
          type: "rounded",
          radius: "0.875em",
        },
      },
      secondaryButton: {
        borderRadius: {
          type: "rounded",
          radius: "0.875em",
        },
      },
      inputField: {
        borderRadius: {
          type: "rounded",
          radius: "0.875em",
        },
      },
      phoneInputField: {
        borderRadius: {
          type: "rounded",
          radius: "0.875em",
        },
      },
      icon: {
        color: "#176df3",
      },
      link: {
        color: "#176df3",
        colorActive: "#1151b8",
        colorHover: "#1151b8",
        textDecoration: "underline",
      },
      logo: {
        height: "2.5rem",
      },
    });
  });

  it("should return default style if target css vars not defined", () => {
    const emptyStyleSheet = "";
    const customisableThemeStyleGroup = new CustomisableThemeStyleGroup(
      DEFAULT_LIGHT_THEME
    );
    const styleVisitor = new StyleCssVisitor(
      ThemeTargetSelector.Light,
      customisableThemeStyleGroup
    );
    const styles = styleVisitor.getStyle(parseCSS(emptyStyleSheet));
    expect(styles).toEqual(DEFAULT_LIGHT_THEME);
  });
});

describe("CssAstVisitor", () => {
  it("should generate css stylesheet", () => {
    const customisableThemeStyleGroup = new CustomisableThemeStyleGroup(
      DEFAULT_LIGHT_THEME
    );
    customisableThemeStyleGroup.setValue({
      page: {
        backgroundColor: "#1c1c1e",
      },
      card: {
        alignment: "center",
      },
      primaryButton: {
        backgroundColor: "#176df3",
        backgroundColorActive: "#235dba",
        backgroundColorHover: "#235dba",
        labelColor: "#1c1c1e",
        borderRadius: {
          type: "rounded",
          radius: "0.875em",
        },
      },
      secondaryButton: {
        borderRadius: {
          type: "rounded",
          radius: "0.875em",
        },
      },
      inputField: {
        borderRadius: {
          type: "rounded",
          radius: "0.875em",
        },
      },
      phoneInputField: {
        borderRadius: {
          type: "rounded",
          radius: "0.875em",
        },
      },
      icon: {
        color: "#2f7bf4",
      },
      link: {
        color: "#2f7bf4",
        colorActive: "#235dba",
        colorHover: "#235dba",
        textDecoration: "underline",
      },
      logo: {},
    });
    const expectedStyleSheet = `:root {
    --layout__bg-color: #1c1c1e;
    --alignment-card: center;
    --primary-btn__bg-color: #176df3;
    --primary-btn__bg-color--active: #235dba;
    --primary-btn__bg-color--hover: #235dba;
    --primary-btn__text-color: #1c1c1e;
    --primary-btn__border-radius: 0.875em;
    --secondary-btn__border-radius: 0.875em;
    --input__border-radius: 0.875em;
    --phone-input__trigger-border-radius: 0.875em;
    --icon-color: #2f7bf4;
    --color-link: #2f7bf4;
    --color-link--active: #235dba;
    --color-link--hover: #235dba;
    --body-text__link-text-decoration: underline;
    --brand-logo__height: 2.5rem
}`;
    const styleVisitor = new CssAstVisitor(ThemeTargetSelector.Light);
    customisableThemeStyleGroup.acceptCssAstVisitor(styleVisitor);
    const generatedStyleSheet = styleVisitor.getCSS().toResult().css;
    expect(generatedStyleSheet).toEqual(expectedStyleSheet);
  });
});
