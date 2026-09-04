import { parseCSSColor, getShadeString, Shade } from "./shades";
import { Root, Node, Rule, Declaration } from "postcss";

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
  paddingTop: string;
  paddingBottom: string;
  paddingLeft: string;
  paddingRight: string;
  backgroundColor: string;
}

export const DEFAULT_BANNER_CONFIGURATION: BannerConfiguration = {
  width: "initial",
  height: "55px",
  paddingTop: "16px",
  paddingRight: "16px",
  paddingBottom: "16px",
  paddingLeft: "16px",
  backgroundColor: "transparent",
};

export function deriveColors(
  color: string
): { original: string; variant: string } | null {
  const colorObject = parseCSSColor(color);
  if (colorObject == null) {
    return null;
  }
  // The variant is FluentUI's themeDark slot: Shade7 of the primary color.
  return {
    original: color,
    variant: getShadeString(colorObject, Shade.Shade7),
  };
}

// This function is for v1 UI. Should avoid using it.
// getShades takes a color and then return the shades.
// The return value is 9-element array, with the first element being the originally given color.
// The remaining 8 elements are the shades, ordered from Shade.Shade1 to Shade.Shade8
export function getShades(colorStr: string): string[] {
  const color = parseCSSColor(colorStr);
  if (color == null) {
    throw new Error("invalid color: " + colorStr);
  }
  // Slot 0 is the original color string verbatim, then Shade1..Shade8.
  return [
    colorStr,
    getShadeString(color, Shade.Shade1),
    getShadeString(color, Shade.Shade2),
    getShadeString(color, Shade.Shade3),
    getShadeString(color, Shade.Shade4),
    getShadeString(color, Shade.Shade5),
    getShadeString(color, Shade.Shade6),
    getShadeString(color, Shade.Shade7),
    getShadeString(color, Shade.Shade8),
  ];
}

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

export function getLightBannerConfiguration(
  nodes: Node[]
): BannerConfiguration | null {
  let width;
  let height;
  let paddingTop;
  let paddingRight;
  let paddingBottom;
  let paddingLeft;
  let backgroundColor;

  for (const rule of nodes) {
    if (rule instanceof Rule) {
      for (const decl of rule.nodes) {
        if (decl instanceof Declaration) {
          if (rule.selector === ".banner") {
            switch (decl.prop) {
              case "width":
                width = decl.value;
                break;
              case "height":
                height = decl.value;
                break;
            }
          }
          if (rule.selector === ".banner-frame") {
            switch (decl.prop) {
              case "padding-top":
                paddingTop = decl.value;
                break;
              case "padding-right":
                paddingRight = decl.value;
                break;
              case "padding-bottom":
                paddingBottom = decl.value;
                break;
              case "padding-left":
                paddingLeft = decl.value;
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

  if (
    width != null &&
    height != null &&
    paddingTop != null &&
    paddingRight != null &&
    paddingBottom != null &&
    paddingLeft != null &&
    backgroundColor != null
  ) {
    return {
      width,
      height,
      paddingTop,
      paddingRight,
      paddingBottom,
      paddingLeft,
      backgroundColor,
    };
  }

  return null;
}

export function getDarkTheme(nodes: Node[]): DarkTheme | null {
  let primaryColor;
  let textColor;
  let backgroundColor;

  for (const rule of nodes) {
    if (rule instanceof Rule && rule.selector === ":root.dark") {
      // Extract theme
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
      isDarkTheme: true,
      primaryColor,
      textColor,
      backgroundColor,
    };
  }

  return null;
}

export function getDarkBannerConfiguration(
  nodes: Node[]
): BannerConfiguration | null {
  let width;
  let height;
  let paddingTop;
  let paddingRight;
  let paddingBottom;
  let paddingLeft;
  let backgroundColor;

  for (const rule of nodes) {
    if (rule instanceof Rule && rule.selector === ".dark .banner-frame") {
      for (const decl of rule.nodes) {
        if (decl instanceof Declaration) {
          switch (decl.prop) {
            case "padding-top":
              paddingTop = decl.value;
              break;
            case "padding-right":
              paddingRight = decl.value;
              break;
            case "padding-bottom":
              paddingBottom = decl.value;
              break;
            case "padding-left":
              paddingLeft = decl.value;
              break;
            case "background-color":
              backgroundColor = decl.value;
              break;
          }
        }
      }
    }

    if (rule instanceof Rule && rule.selector === ".dark .banner") {
      for (const decl of rule.nodes) {
        if (decl instanceof Declaration) {
          switch (decl.prop) {
            case "width":
              width = decl.value;
              break;
            case "height":
              height = decl.value;
              break;
          }
        }
      }
    }
  }

  if (
    width != null &&
    height != null &&
    paddingTop != null &&
    paddingRight != null &&
    paddingBottom != null &&
    paddingLeft != null &&
    backgroundColor != null
  ) {
    return {
      width,
      height,
      paddingTop,
      paddingRight,
      paddingBottom,
      paddingLeft,
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

export function addLightTheme(root: Root, lightTheme: LightTheme): void {
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
}

export function addDarkTheme(root: Root, darkTheme: DarkTheme): void {
  const darkPseudoRoot = new Rule({ selector: ":root.dark" });
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
  root.append(darkPseudoRoot);
}

export function addLightBannerConfiguration(
  root: Root,
  c: BannerConfiguration
): void {
  const bannerRule = new Rule({ selector: ".banner" });
  bannerRule.append(new Declaration({ prop: "width", value: c.width }));
  bannerRule.append(new Declaration({ prop: "height", value: c.height }));

  const bannerFrameRule = new Rule({ selector: ".banner-frame" });
  bannerFrameRule.append(
    new Declaration({ prop: "padding-top", value: c.paddingTop })
  );
  bannerFrameRule.append(
    new Declaration({ prop: "padding-right", value: c.paddingRight })
  );
  bannerFrameRule.append(
    new Declaration({ prop: "padding-bottom", value: c.paddingBottom })
  );
  bannerFrameRule.append(
    new Declaration({ prop: "padding-left", value: c.paddingLeft })
  );
  bannerFrameRule.append(
    new Declaration({ prop: "background-color", value: c.backgroundColor })
  );

  root.append(bannerRule);
  root.append(bannerFrameRule);
}

export function addDarkBannerConfiguration(
  root: Root,
  c: BannerConfiguration
): void {
  const bannerRule = new Rule({ selector: ".dark .banner" });
  bannerRule.append(new Declaration({ prop: "width", value: c.width }));
  bannerRule.append(new Declaration({ prop: "height", value: c.height }));

  const bannerFrameRule = new Rule({ selector: ".dark .banner-frame" });
  bannerFrameRule.append(
    new Declaration({ prop: "padding-top", value: c.paddingTop })
  );
  bannerFrameRule.append(
    new Declaration({ prop: "padding-right", value: c.paddingRight })
  );
  bannerFrameRule.append(
    new Declaration({ prop: "padding-bottom", value: c.paddingBottom })
  );
  bannerFrameRule.append(
    new Declaration({ prop: "padding-left", value: c.paddingLeft })
  );
  bannerFrameRule.append(
    new Declaration({ prop: "background-color", value: c.backgroundColor })
  );

  root.append(bannerRule);
  root.append(bannerFrameRule);
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
