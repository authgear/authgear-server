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

export interface BannerConfiguration {
  width: string;
  height: string;
  marginTop: string;
  marginBottom: string;
  marginLeft: string;
  marginRight: string;
  backgroundColor: string;
}

export const DEFAULT_BANNER_CONFIGURATION: BannerConfiguration = {
  width: "initial",
  height: "55px",
  marginTop: "16px",
  marginRight: "16px",
  marginBottom: "16px",
  marginLeft: "16px",
  backgroundColor: "transparent",
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

// eslint-disable-next-line complexity
export function getLightTheme(nodes: Node[]): LightTheme | null {
  let primaryColor;
  let textColor;
  let backgroundColor;

  for (const rule of nodes) {
    if (rule instanceof Rule && rule.selector === ":root") {
      for (const decl of rule.nodes) {
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
export function getLightBannerConfiguration(
  nodes: Node[]
): BannerConfiguration | null {
  let width;
  let height;
  let marginTop;
  let marginRight;
  let marginBottom;
  let marginLeft;
  let backgroundColor;

  for (const rule of nodes) {
    if (rule instanceof Rule && rule.selector === ".banner") {
      for (const decl of rule.nodes) {
        if (decl instanceof Declaration) {
          switch (decl.prop) {
            case "width":
              width = decl.value;
              break;
            case "height":
              height = decl.value;
              break;
            case "margin-top":
              marginTop = decl.value;
              break;
            case "margin-right":
              marginRight = decl.value;
              break;
            case "margin-bottom":
              marginBottom = decl.value;
              break;
            case "margin-left":
              marginLeft = decl.value;
              break;
            case "background-color":
              backgroundColor = decl.value;
              break;
          }
        }
      }
    }
  }

  if (
    width != null &&
    height != null &&
    marginTop != null &&
    marginRight != null &&
    marginBottom != null &&
    marginLeft != null &&
    backgroundColor != null
  ) {
    return {
      width,
      height,
      marginTop,
      marginRight,
      marginBottom,
      marginLeft,
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
      for (const rule of atRule.nodes) {
        // Extract theme
        if (rule instanceof Rule && rule.selector === ":root") {
          for (const decl of rule.nodes) {
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

// eslint-disable-next-line complexity
export function getDarkBannerConfiguration(
  nodes: Node[]
): BannerConfiguration | null {
  let width;
  let height;
  let marginTop;
  let marginRight;
  let marginBottom;
  let marginLeft;
  let backgroundColor;

  for (const atRule of nodes) {
    if (
      atRule instanceof AtRule &&
      atRule.params === "(prefers-color-scheme: dark)"
    ) {
      for (const rule of atRule.nodes) {
        if (rule instanceof Rule && rule.selector === ".banner") {
          for (const decl of rule.nodes) {
            if (decl instanceof Declaration) {
              switch (decl.prop) {
                case "width":
                  width = decl.value;
                  break;
                case "height":
                  height = decl.value;
                  break;
                case "margin-top":
                  marginTop = decl.value;
                  break;
                case "margin-right":
                  marginRight = decl.value;
                  break;
                case "margin-bottom":
                  marginBottom = decl.value;
                  break;
                case "margin-left":
                  marginLeft = decl.value;
                  break;
                case "background-color":
                  backgroundColor = decl.value;
                  break;
              }
            }
          }
        }
      }
    }
  }

  if (
    width != null &&
    height != null &&
    marginTop != null &&
    marginRight != null &&
    marginBottom != null &&
    marginLeft != null &&
    backgroundColor != null
  ) {
    return {
      width,
      height,
      marginTop,
      marginRight,
      marginBottom,
      marginLeft,
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
