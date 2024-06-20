import { Declaration, Root } from "postcss";
import {
  CssDeclarationNodeWrapper,
  CssNodeVisitor,
  CssOtherNodeWrapper,
  CssRootNodeWrapper,
  CssRuleNodeWrapper,
} from "../util/cssVisitor";

type Color = string;

export type Alignment = "start" | "center" | "end";

export type BorderRadiusStyleType = "none" | "rounded" | "rounded-full";

export type BorderRadiusStyle =
  | {
      type: "none";
    }
  | {
      type: "rounded";
      radius: string;
    }
  | {
      type: "rounded-full";
    };

export interface ButtonStyle {
  backgroundColor: Color;
  labelColor: Color;
  borderRadius: BorderRadiusStyle;
}

export interface InputFieldStyle {
  borderRadius: BorderRadiusStyle;
}

export interface LinkStyle {
  color: Color;
}

export interface CustomisableTheme {
  cardAlignment: Alignment;
  backgroundColor: Color;

  primaryButton: ButtonStyle;
  inputField: InputFieldStyle;

  link: LinkStyle;
}

export const DEFAULT_LIGHT_THEME: CustomisableTheme = {
  cardAlignment: "center",
  backgroundColor: "#ffffff",

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
};

abstract class AbstractStyle<T> {
  abstract acceptDeclaration(declaration: Declaration): boolean;
  abstract getValue(): T;
}

abstract class StyleProperty<T> extends AbstractStyle<T> {
  private readonly propertyName: string;
  value: T;

  constructor(propertyName: string, defaultValue: T) {
    super();
    this.propertyName = propertyName;
    this.value = defaultValue;
  }

  abstract setValue(rawValue: string): void;

  acceptDeclaration(declaration: Declaration): boolean {
    if (declaration.prop !== this.propertyName) {
      return false;
    }
    this.setValue(declaration.value);
    return true;
  }

  getValue(): T {
    return this.value;
  }
}

class ColorStyleProperty extends StyleProperty<string> {
  setValue(rawValue: string): void {
    this.value = rawValue;
  }
}

class AlignItemsStyleProperty extends StyleProperty<Alignment> {
  setValue(rawValue: string): void {
    switch (rawValue) {
      case "flex-start":
        this.value = "start";
        break;
      case "flex-end":
        this.value = "end";
        break;
      default:
        this.value = "center";
        break;
    }
  }
}

class BorderRadiusStyleProperty extends StyleProperty<BorderRadiusStyle> {
  setValue(rawValue: string): void {
    switch (rawValue) {
      case "9999px":
        this.value = {
          type: "rounded-full",
        };
        break;
      case "0":
        this.value = {
          type: "none",
        };
        break;
      default:
        this.value = {
          type: "rounded",
          radius: rawValue,
        };
        break;
    }
  }
}

type StyleProperties<T> = {
  [K in keyof T]: AbstractStyle<T[K] | null>;
};
class StyleGroup<T> extends AbstractStyle<T> {
  private styles: StyleProperties<T>;
  constructor(styles: StyleProperties<T>) {
    super();
    this.styles = styles;
  }

  acceptDeclaration(declaration: Declaration): boolean {
    for (const style of Object.values(this.styles)) {
      const s = style as AbstractStyle<T>;
      if (s.acceptDeclaration(declaration)) {
        return true;
      }
    }
    return false;
  }

  getValue(): T {
    const value: Record<string, unknown> = {};
    for (const [name, style] of Object.entries(this.styles)) {
      const s = style as AbstractStyle<T>;
      value[name] = s.getValue();
    }
    return value as T;
  }
}

export class CustomisableThemeStyleGroup extends StyleGroup<CustomisableTheme> {
  constructor() {
    super({
      cardAlignment: new AlignItemsStyleProperty(
        "--layout-flex-align-items",
        DEFAULT_LIGHT_THEME.cardAlignment
      ),
      backgroundColor: new ColorStyleProperty(
        "-—widget__bg-color",
        DEFAULT_LIGHT_THEME.backgroundColor
      ),

      primaryButton: new StyleGroup({
        backgroundColor: new ColorStyleProperty(
          "-—primary-btn__bg-color",
          DEFAULT_LIGHT_THEME.primaryButton.backgroundColor
        ),
        labelColor: new ColorStyleProperty(
          "—-primary-btn__text-color",
          DEFAULT_LIGHT_THEME.primaryButton.labelColor
        ),
        borderRadius: new BorderRadiusStyleProperty(
          "—-primary-btn__border-radius",
          DEFAULT_LIGHT_THEME.primaryButton.borderRadius
        ),
      }),

      inputField: new StyleGroup({
        borderRadius: new BorderRadiusStyleProperty(
          "--input__border-radius",
          DEFAULT_LIGHT_THEME.inputField.borderRadius
        ),
      }),

      link: new StyleGroup({
        color: new ColorStyleProperty(
          "--body-text__link-color",
          DEFAULT_LIGHT_THEME.link.color
        ),
      }),
    });
  }
}

export class StyleCSSVisitor<T> extends CssNodeVisitor {
  private ruleSelector: string;

  private styleGroup: StyleGroup<T>;

  constructor(ruleSelector: string, styleGroup: StyleGroup<T>) {
    super();
    this.ruleSelector = ruleSelector;
    this.styleGroup = styleGroup;
  }

  visitRoot(root: CssRootNodeWrapper): void {
    root.nodes.forEach((node) => {
      node.accept(this);
    });
  }

  visitRule(rule: CssRuleNodeWrapper): void {
    if (rule.selector !== this.ruleSelector) {
      return;
    }
    rule.nodes.forEach((node) => {
      node.accept(this);
    });
  }

  visitDeclaration(declaration: CssDeclarationNodeWrapper): void {
    this.styleGroup.acceptDeclaration(declaration.declaration);
  }

  // eslint-disable-next-line
  visitOther(_: CssOtherNodeWrapper): void {
    // no-op
  }

  getStyle(root: Root): T {
    const wrapper = new CssRootNodeWrapper(root);
    wrapper.accept(this);
    return this.styleGroup.getValue();
  }
}
