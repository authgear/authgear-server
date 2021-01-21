import {
  getColorFromString,
  themeRulesStandardCreator,
  ThemeGenerator,
  BaseSlots,
} from "@fluentui/react";
import { Node, Rule, AtRule, Declaration } from "postcss";

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

export interface Theme {
  lightModePrimaryColor: string;
  lightModeTextColor: string;
  lightModeBackgroundColor: string;
  darkModePrimaryColor: string;
  darkModeTextColor: string;
  darkModeBackgroundColor: string;
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
