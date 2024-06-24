import { Declaration, Root, Rule } from "postcss";
import {
  CssDeclarationNodeWrapper,
  CssNodeVisitor,
  CssOtherNodeWrapper,
  CssRootNodeWrapper,
  CssRuleNodeWrapper,
} from "../util/cssVisitor";

export const enum ThemeTargetSelector {
  Light = ":root",
}

type Color = string;

export const AllAlignments = ["start", "center", "end"] as const;
export type Alignment = typeof AllAlignments[number];

export const AllBorderRadiusStyleTypes = [
  "none",
  "rounded",
  "rounded-full",
] as const;
export type BorderRadiusStyleType = typeof AllBorderRadiusStyleTypes[number];

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
  abstract acceptCssAstVisitor(visitor: CssAstVisitor): void;
  abstract getValue(): T;
}

abstract class StyleProperty<T> extends AbstractStyle<T> {
  readonly propertyName: string;
  value: T;

  constructor(propertyName: string, defaultValue: T) {
    super();
    this.propertyName = propertyName;
    this.value = defaultValue;
  }

  protected abstract setWithRawValue(rawValue: string): void;

  acceptDeclaration(declaration: Declaration): boolean {
    if (declaration.prop !== this.propertyName) {
      return false;
    }
    this.setWithRawValue(declaration.value);
    return true;
  }

  getValue(): T {
    return this.value;
  }

  abstract getCSSValue(): string | number;
}

class ColorStyleProperty extends StyleProperty<string> {
  protected setWithRawValue(rawValue: string): void {
    this.value = rawValue;
  }

  acceptCssAstVisitor(visitor: CssAstVisitor): void {
    visitor.visitColorStyleProperty(this);
  }

  getCSSValue(): string {
    return this.value;
  }
}

class AlignItemsStyleProperty extends StyleProperty<Alignment> {
  protected setWithRawValue(rawValue: string): void {
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

  acceptCssAstVisitor(visitor: CssAstVisitor): void {
    visitor.visitAlignItemsStyleProperty(this);
  }

  getCSSValue(): string {
    switch (this.value) {
      case "start":
        return "flex-start";
      case "end":
        return "flex-end";
      case "center":
        return "center";
      default:
        return "";
    }
  }
}

class BorderRadiusStyleProperty extends StyleProperty<BorderRadiusStyle> {
  static FULL_ROUNDED_CSS_VALUE = "9999px";

  protected setWithRawValue(rawValue: string): void {
    switch (rawValue) {
      case BorderRadiusStyleProperty.FULL_ROUNDED_CSS_VALUE:
        this.value = {
          type: "rounded-full",
        };
        break;
      case "initial":
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

  acceptCssAstVisitor(visitor: CssAstVisitor): void {
    visitor.visitBorderRadiusStyleProperty(this);
  }

  getCSSValue(): string {
    switch (this.value.type) {
      case "rounded":
        return this.value.radius;
      case "rounded-full":
        return BorderRadiusStyleProperty.FULL_ROUNDED_CSS_VALUE;
      case "none":
        return "initial";
      default:
        return "";
    }
  }
}

type StyleProperties<T> = {
  [K in keyof T]: AbstractStyle<T[K] | null>;
};
class StyleGroup<T> extends AbstractStyle<T> {
  styles: StyleProperties<T>;

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

  acceptCssAstVisitor(visitor: CssAstVisitor): void {
    visitor.visitStyleGroup(this);
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
  constructor(value: CustomisableTheme = DEFAULT_LIGHT_THEME) {
    super({
      cardAlignment: new AlignItemsStyleProperty(
        "--layout-flex-align-items",
        value.cardAlignment
      ),
      backgroundColor: new ColorStyleProperty(
        "-—widget__bg-color",
        value.backgroundColor
      ),

      primaryButton: new StyleGroup({
        backgroundColor: new ColorStyleProperty(
          "-—primary-btn__bg-color",
          value.primaryButton.backgroundColor
        ),
        labelColor: new ColorStyleProperty(
          "—-primary-btn__text-color",
          value.primaryButton.labelColor
        ),
        borderRadius: new BorderRadiusStyleProperty(
          "—-primary-btn__border-radius",
          value.primaryButton.borderRadius
        ),
      }),

      inputField: new StyleGroup({
        borderRadius: new BorderRadiusStyleProperty(
          "--input__border-radius",
          value.inputField.borderRadius
        ),
      }),

      link: new StyleGroup({
        color: new ColorStyleProperty(
          "--body-text__link-color",
          value.link.color
        ),
      }),
    });
  }
}

export class StyleCssVisitor<T> extends CssNodeVisitor {
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

export class CssAstVisitor {
  private root: Root;
  private rule: Rule;

  constructor(ruleSelector: string) {
    this.root = new Root();
    this.rule = new Rule({
      selector: ruleSelector,
    });
    this.root.append(this.rule);
  }

  visitStyleGroup<T>(styleGroup: StyleGroup<T>): void {
    for (const style of Object.values(styleGroup.styles)) {
      const s = style as AbstractStyle<T>;
      s.acceptCssAstVisitor(this);
    }
  }

  visitAlignItemsStyleProperty(styleProperty: AlignItemsStyleProperty): void {
    this.visitorStyleProperty(styleProperty);
  }

  visitBorderRadiusStyleProperty(
    styleProperty: BorderRadiusStyleProperty
  ): void {
    this.visitorStyleProperty(styleProperty);
  }

  visitColorStyleProperty(styleProperty: ColorStyleProperty): void {
    this.visitorStyleProperty(styleProperty);
  }

  visitorStyleProperty<T>(styleProperty: StyleProperty<T>): void {
    this.rule.append(
      new Declaration({
        prop: styleProperty.propertyName,
        value: String(styleProperty.getCSSValue()),
      })
    );
  }

  getCSS(): Root {
    return this.root;
  }
}
