import {
  getColorFromString,
  themeRulesStandardCreator,
  ThemeGenerator,
  BaseSlots,
} from "@fluentui/react";
import { Root, Node, Rule, AtRule, Declaration } from "postcss";

export interface Theme {
  lightModePrimaryColor: string;
  lightModeTextColor: string;
  lightModeBackgroundColor: string;
  darkModePrimaryColor: string;
  darkModeTextColor: string;
  darkModeBackgroundColor: string;
}

export const THEME_DELIMITER_COMMENT = "AUTHGEAR THEME CSS. DO NOT EDIT!";

export const DEFAULT_THEME: Theme = {
  lightModePrimaryColor: "#176df3",
  lightModeTextColor: "#000000",
  lightModeBackgroundColor: "#ffffff",
  darkModePrimaryColor: "#317BF4",
  darkModeTextColor: "#ffffff",
  darkModeBackgroundColor: "#000000",
};

// getShades takes a color and then return the shades.
// The return value is 9-element array, with the first element being the originally given color.
// The remaining 8 elements are the shades, ordered from Shade.Shade1 to Shade.Shade8
export function getShades(colorStr: string): string[] {
  const themeRules = themeRulesStandardCreator();
  const color = getColorFromString(colorStr);
  if (color == null) {
    throw new Error("invalid color: " + colorStr);
  }
  ThemeGenerator.insureSlots(themeRules, false);
  // It is extremely important to pass trailing (true, true) to setSlot,
  // otherwise setSlot does not take effect at all.
  ThemeGenerator.setSlot(
    themeRules[BaseSlots[BaseSlots.primaryColor]],
    color,
    false,
    true,
    true
  );

  const json = ThemeGenerator.getThemeAsJson(themeRules);
  const {
    primaryColor,
    primaryColorShade1,
    primaryColorShade2,
    primaryColorShade3,
    primaryColorShade4,
    primaryColorShade5,
    primaryColorShade6,
    primaryColorShade7,
    primaryColorShade8,
  } = json;

  return [
    primaryColor,
    primaryColorShade1,
    primaryColorShade2,
    primaryColorShade3,
    primaryColorShade4,
    primaryColorShade5,
    primaryColorShade6,
    primaryColorShade7,
    primaryColorShade8,
  ];
}
// getTheme takes a list of CSS nodes and extract the theme.
// eslint-disable-next-line complexity
export function getTheme(nodes: Node[]): Theme | null {
  let lightModePrimaryColor;
  let lightModeTextColor;
  let lightModeBackgroundColor;
  let darkModePrimaryColor;
  let darkModeTextColor;
  let darkModeBackgroundColor;

  // Extract light mode.
  for (const pseudoRoot of nodes) {
    if (pseudoRoot instanceof Rule && pseudoRoot.selector === ":root") {
      for (const decl of pseudoRoot.nodes) {
        if (decl instanceof Declaration) {
          switch (decl.prop) {
            case "--color-primary-unshaded":
              lightModePrimaryColor = decl.value;
              break;
            case "--color-text-unshaded":
              lightModeTextColor = decl.value;
              break;
            case "--color-background-unshaded":
              lightModeBackgroundColor = decl.value;
              break;
          }
        }
      }
    }
  }

  // Extract dark mode.
  for (const atRule of nodes) {
    if (
      atRule instanceof AtRule &&
      atRule.params === "(prefers-color-scheme: dark)"
    ) {
      for (const pseudoRoot of atRule.nodes) {
        if (pseudoRoot instanceof Rule && pseudoRoot.selector === ":root") {
          for (const decl of pseudoRoot.nodes) {
            if (decl instanceof Declaration) {
              switch (decl.prop) {
                case "--color-primary-unshaded":
                  darkModePrimaryColor = decl.value;
                  break;
                case "--color-text-unshaded":
                  darkModeTextColor = decl.value;
                  break;
                case "--color-background-unshaded":
                  darkModeBackgroundColor = decl.value;
                  break;
              }
            }
          }
        }
      }
    }
  }

  if (
    lightModePrimaryColor != null &&
    lightModeTextColor != null &&
    lightModeBackgroundColor != null &&
    darkModePrimaryColor != null &&
    darkModeTextColor != null &&
    darkModeBackgroundColor != null
  ) {
    return {
      lightModePrimaryColor,
      lightModeTextColor,
      lightModeBackgroundColor,
      darkModePrimaryColor,
      darkModeTextColor,
      darkModeBackgroundColor,
    };
  }

  return null;
}

function addShadeDeclarations(rule: Rule, shades: string[], name: string) {
  for (let i = 0; i < shades.length; i++) {
    const value = shades[i];
    if (i === 0) {
      rule.append(new Declaration({ prop: `--color-${name}-unshaded`, value }));
    } else {
      rule.append(
        new Declaration({ prop: `--color-${name}-shaded-${i}`, value })
      );
    }
  }
}

export function themeToCSS(theme: Theme): string {
  const root = new Root();

  const pseudoRoot = new Rule({ selector: ":root" });
  addShadeDeclarations(
    pseudoRoot,
    getShades(theme.lightModePrimaryColor),
    "primary"
  );
  addShadeDeclarations(pseudoRoot, getShades(theme.lightModeTextColor), "text");
  addShadeDeclarations(
    pseudoRoot,
    getShades(theme.lightModeBackgroundColor),
    "background"
  );
  root.append(pseudoRoot);

  const atRule = new AtRule({
    name: "media",
    params: "(prefers-color-scheme: dark)",
  });
  const darkPseudoRoot = new Rule({ selector: ":root" });
  addShadeDeclarations(
    darkPseudoRoot,
    getShades(theme.darkModePrimaryColor),
    "primary"
  );
  addShadeDeclarations(
    darkPseudoRoot,
    getShades(theme.darkModeTextColor),
    "text"
  );
  addShadeDeclarations(
    darkPseudoRoot,
    getShades(theme.darkModeBackgroundColor),
    "background"
  );
  atRule.append(darkPseudoRoot);
  root.append(atRule);

  return root.toResult().css;
}
