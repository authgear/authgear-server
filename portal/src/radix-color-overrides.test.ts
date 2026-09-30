import { describe, expect, it } from "@jest/globals";
import * as fs from "fs";
import * as path from "path";
import { parse, Rule, ChildNode } from "postcss";

// Guards the custom palette against a silent dark-mode failure.
//
// The light blocks are selected by `:root`, which matches <html> whether or not
// it carries the `dark` class, and `:root` has the same specificity as `.dark`.
// This file is imported after @radix-ui/themes/styles.css, so a token declared
// only for light overrides Radix's own dark value and stays light in dark mode.
// Every token therefore needs a counterpart in a `.dark` block here.

const CSS = fs.readFileSync(
  path.join(__dirname, "radix-color-overrides.css"),
  "utf8"
);

const LIGHT_SELECTORS = [":root", ".light", ".light-theme"];
const DARK_SELECTORS = [".dark", ".dark-theme"];

function selectorsOf(rule: Rule): string[] {
  return rule.selector.split(",").map((s) => s.trim());
}

function isLight(rule: Rule): boolean {
  return selectorsOf(rule).some((s) => LIGHT_SELECTORS.includes(s));
}

function isDark(rule: Rule): boolean {
  return selectorsOf(rule).some((s) => DARK_SELECTORS.includes(s));
}

/**
 * Custom property names declared by light and by dark rules. `insideSupports`
 * picks the wide-gamut (`@supports` + `@media (color-gamut: p3)`) copies
 * instead of the plain ones, which are declared separately and can drift apart.
 */
function collect(insideSupports: boolean): { light: string[]; dark: string[] } {
  const root = parse(CSS);
  const light = new Set<string>();
  const dark = new Set<string>();

  const visit = (node: ChildNode, underSupports: boolean) => {
    if (node.type === "atrule") {
      const atRule = node;
      const nowUnder = underSupports || atRule.name === "supports";
      atRule.each((child) => visit(child, nowUnder));
      return;
    }
    if (node.type !== "rule") {
      return;
    }
    const rule = node;
    if (underSupports !== insideSupports) {
      return;
    }
    const target = isLight(rule) ? light : isDark(rule) ? dark : null;
    if (target == null) {
      return;
    }
    rule.each((decl) => {
      if (decl.type === "decl" && decl.prop.startsWith("--")) {
        target.add(decl.prop);
      }
    });
  };

  root.each((node) => visit(node, false));
  return { light: [...light].sort(), dark: [...dark].sort() };
}

describe("radix-color-overrides.css", () => {
  it("declares a dark counterpart for every light token", () => {
    const { light, dark } = collect(false);
    expect(light.length).toBeGreaterThan(0);
    expect(dark).toEqual(light);
  });

  it("declares a dark counterpart for every wide-gamut light token", () => {
    const { light, dark } = collect(true);
    expect(light.length).toBeGreaterThan(0);
    expect(dark).toEqual(light);
  });
});
