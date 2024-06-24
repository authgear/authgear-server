import { describe, it, expect } from "@jest/globals";

import { Declaration, Rule, parse as parseCSS } from "postcss";
import {
  ColorStyleProperty,
  CssAstVisitor,
  StyleCssVisitor,
  StyleGroup,
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
});

describe("CssAstVisitor", () => {
  it("should generate css stylesheet", () => {
    const styleGroup = new StyleGroup({
      button: new StyleGroup({
        textColor: new ColorStyleProperty("-—primary-btn__bg-color", "red"),
      }),
    });

    const styleVisitor = new CssAstVisitor(":root");
    styleGroup.acceptCssAstVisitor(styleVisitor);
    const css = styleVisitor.getCSS();
    expect(css.nodes.length).toEqual(1);
    expect(css.nodes[0].type).toEqual("rule");

    const rule = css.nodes[0] as Rule;
    expect(rule.nodes.length).toEqual(1);
    expect(rule.nodes[0].type).toEqual("decl");

    const node = rule.nodes[0] as Declaration;
    expect(node.prop).toEqual("-—primary-btn__bg-color");
    expect(node.value).toEqual("red");
  });
});
