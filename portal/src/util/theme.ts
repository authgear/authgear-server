import {
  getColorFromString,
  themeRulesStandardCreator,
  ThemeGenerator,
  BaseSlots,
} from "@fluentui/react";
import { Root, Node, Rule, AtRule, Declaration } from "postcss";

export interface LightTheme {
  isDarkTheme: false;
  primaryColor: string;
  textColor: string;
  backgroundColor: string;
}

export interface DarkTheme {
  isDarkTheme: true;
  primaryColor: string;
  textColor: string;
  backgroundColor: string;
}

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

// eslint-disable-next-line complexity
export function getLightTheme(nodes: Node[]): LightTheme | null {
  let primaryColor;
  let textColor;
  let backgroundColor;

  for (const pseudoRoot of nodes) {
    if (pseudoRoot instanceof Rule && pseudoRoot.selector === ":root") {
      for (const decl of pseudoRoot.nodes) {
        if (decl instanceof Declaration) {
          switch (decl.prop) {
            case "--color-primary-unshaded":
              primaryColor = decl.value;
              break;
            case "--color-text-unshaded":
              textColor = decl.value;
              break;
            case "--color-background-unshaded":
              backgroundColor = decl.value;
              break;
          }
        }
      }
    }
  }

  if (primaryColor != null && textColor != null && backgroundColor != null) {
    return {
      isDarkTheme: false,
      primaryColor,
      textColor,
      backgroundColor,
    };
  }

  return null;
}

// eslint-disable-next-line complexity
export function getDarkTheme(nodes: Node[]): DarkTheme | null {
  let primaryColor;
  let textColor;
  let backgroundColor;

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
                  primaryColor = decl.value;
                  break;
                case "--color-text-unshaded":
                  textColor = decl.value;
                  break;
                case "--color-background-unshaded":
                  backgroundColor = decl.value;
                  break;
              }
            }
          }
        }
      }
    }
  }

  if (primaryColor != null && textColor != null && backgroundColor != null) {
    return {
      isDarkTheme: true,
      primaryColor,
      textColor,
      backgroundColor,
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

export function lightThemeToCSS(lightTheme: LightTheme): string {
  const root = new Root();

  const pseudoRoot = new Rule({ selector: ":root" });
  addShadeDeclarations(
    pseudoRoot,
    getShades(lightTheme.primaryColor),
    "primary"
  );
  addShadeDeclarations(pseudoRoot, getShades(lightTheme.textColor), "text");
  addShadeDeclarations(
    pseudoRoot,
    getShades(lightTheme.backgroundColor),
    "background"
  );
  root.append(pseudoRoot);

  return root.toResult().css;
}

export function darkThemeToCSS(darkTheme: DarkTheme): string {
  const root = new Root();

  const atRule = new AtRule({
    name: "media",
    params: "(prefers-color-scheme: dark)",
  });
  const darkPseudoRoot = new Rule({ selector: ":root" });
  addShadeDeclarations(
    darkPseudoRoot,
    getShades(darkTheme.primaryColor),
    "primary"
  );
  addShadeDeclarations(darkPseudoRoot, getShades(darkTheme.textColor), "text");
  addShadeDeclarations(
    darkPseudoRoot,
    getShades(darkTheme.backgroundColor),
    "background"
  );
  atRule.append(darkPseudoRoot);
  root.append(atRule);

  return root.toResult().css;
}

export function isLightThemeEqual(a: LightTheme, b: LightTheme): boolean {
  return (
    a.primaryColor === b.primaryColor &&
    a.textColor === b.textColor &&
    a.backgroundColor === b.backgroundColor
  );
}

export function isDarkThemeEqual(a: DarkTheme, b: DarkTheme): boolean {
  return (
    a.primaryColor === b.primaryColor &&
    a.textColor === b.textColor &&
    a.backgroundColor === b.backgroundColor
  );
}
