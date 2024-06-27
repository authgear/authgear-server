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
    --layout-flex-align-items: center;
    --layout-padding-left: 0;
    --layout-padding-right: 0,;
    --primary-btn__bg-color: #176df3;
    --primary-btn__text-color: #ffffff;
    --primary-btn__border-radius: 0.875em;
    --input__border-radius: 0.875em;
    --body-text__link-color: #176df3
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
        leftMargin: "0",
        rightMargin: "0,",
      },
      primaryButton: {
        backgroundColor: "#176df3",
        labelColor: "#ffffff",
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
      link: {
        color: "#176df3",
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
        leftMargin: "0",
        rightMargin: "0,",
      },
      primaryButton: {
        backgroundColor: "#176df3",
        labelColor: "#1c1c1e",
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
      link: {
        color: "#2f7bf4",
      },
    });
    const expectedStyleSheet = `:root {
    --layout__bg-color: #1c1c1e;
    --layout-flex-align-items: center;
    --layout-padding-left: 0;
    --layout-padding-right: 0,;
    --primary-btn__bg-color: #176df3;
    --primary-btn__text-color: #1c1c1e;
    --primary-btn__border-radius: 0.875em;
    --input__border-radius: 0.875em;
    --body-text__link-color: #2f7bf4
}`;
    const styleVisitor = new CssAstVisitor(ThemeTargetSelector.Light);
    customisableThemeStyleGroup.acceptCssAstVisitor(styleVisitor);
    const generatedStyleSheet = styleVisitor.getCSS().toResult().css;
    expect(generatedStyleSheet).toEqual(expectedStyleSheet);
  });
});
