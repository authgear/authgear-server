import { Declaration, Root, Rule } from "postcss";
import {
  CssDeclarationNodeWrapper,
  CssNodeVisitor,
  CssOtherNodeWrapper,
  CssRootNodeWrapper,
  CssRuleNodeWrapper,
} from "../util/cssVisitor";

export enum Theme {
  Light = "light",
  Dark = "dark",
}

export const enum ThemeTargetSelector {
  Light = ":root",
  Dark = ":root.dark",
}
export function getThemeTargetSelector(theme: Theme): ThemeTargetSelector {
  switch (theme) {
    case Theme.Light:
      return ThemeTargetSelector.Light;
    case Theme.Dark:
      return ThemeTargetSelector.Dark;
    default:
      return ThemeTargetSelector.Light;
  }
}

export function selectByTheme<T>(option: { [t in Theme]: T }, theme: Theme): T {
  return option[theme];
}

export const enum CSSVariable {
  AlignmentCard = "--alignment-card",
  LayoutBackgroundColor = "--layout__bg-color",
  LayoutBackgroundImage = "--layout__bg-image",
  PrimaryButtonBackgroundColor = "--primary-btn__bg-color",
  PrimaryButtonBackgroundColorHover = "--primary-btn__bg-color--hover",
  PrimaryButtonBackgroundColorActive = "--primary-btn__bg-color--active",
  PrimaryButtonTextColor = "--primary-btn__text-color",
  PrimaryButtonBorderRadius = "--primary-btn__border-radius",
  SecondaryButtonBorderRadius = "--secondary-btn__border-radius",
  InputFiledBorderRadius = "--input__border-radius",
  LinkColor = "--body-text__link-color",
  WatermarkDisplay = "--watermark-display",
  LogoHeight = "--brand-logo__height",
  PhoneInputTriggerBorderRadius = "--phone-input__trigger-border-radius",
}

export type CSSColor = string;

export const AllAlignments = ["start", "center", "end"] as const;
export type Alignment = (typeof AllAlignments)[number];

export type Hidden = "hidden";

export const AllBorderRadiusStyleTypes = [
  "none",
  "rounded",
  "rounded-full",
] as const;
export type BorderRadiusStyleType = (typeof AllBorderRadiusStyleTypes)[number];

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
  backgroundColor: CSSColor;
  backgroundColorActive: CSSColor;
  backgroundColorHover: CSSColor;
  labelColor: CSSColor;
  borderRadius: BorderRadiusStyle;
}

export interface SecondaryButtonStyle {
  borderRadius: BorderRadiusStyle;
}

export interface PageStyle {
  backgroundColor: CSSColor;
}

export interface CardStyle {
  alignment: Alignment;
}

export interface InputFieldStyle {
  borderRadius: BorderRadiusStyle;
}

export interface PhoneInputFieldStyle {
  borderRadius: BorderRadiusStyle;
}

export interface LinkStyle {
  color: CSSColor;
}

export interface LogoStyte {
  height: string;
}

export const WatermarkEnabledDisplay = "inline-block";
export const WatermarkDisabledDisplay = "hidden";

export interface CustomisableTheme {
  page: PageStyle;
  card: CardStyle;
  primaryButton: ButtonStyle;
  secondaryButton: SecondaryButtonStyle;
  inputField: InputFieldStyle;
  phoneInputField: PhoneInputFieldStyle;
  link: LinkStyle;
  logo: LogoStyte;
}

export interface PartialCustomisableTheme {
  page: Partial<PageStyle>;
  card: Partial<CardStyle>;
  primaryButton: Partial<ButtonStyle>;
  secondaryButton: Partial<SecondaryButtonStyle>;
  inputField: Partial<InputFieldStyle>;
  phoneInputField: Partial<PhoneInputFieldStyle>;
  link: Partial<LinkStyle>;
  logo: Partial<LogoStyte>;
}

export const EMPTY_THEME: PartialCustomisableTheme = {
  page: {},
  card: {},
  primaryButton: {},
  secondaryButton: {},
  inputField: {},
  phoneInputField: {},
  link: {},
  logo: {},
};

export const DEFAULT_BORDER_RADIUS = "0.875em";

export const DEFAULT_LIGHT_THEME: CustomisableTheme = {
  page: {
    backgroundColor: "#ffffff",
  },
  card: {
    alignment: "center",
  },
  primaryButton: {
    backgroundColor: "#176df3",
    backgroundColorActive: "#1151b8",
    backgroundColorHover: "#1151b8",
    labelColor: "#ffffff",
    borderRadius: {
      type: "rounded",
      radius: "0.875em",
    },
  },
  secondaryButton: {
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
  phoneInputField: {
    borderRadius: {
      type: "rounded",
      radius: "0.875em",
    },
  },
  link: {
    color: "#176df3",
  },
  logo: {
    height: "2.5rem",
  },
};

export const DEFAULT_DARK_THEME: CustomisableTheme = {
  page: {
    backgroundColor: "#1c1c1e",
  },
  card: {
    alignment: "center",
  },
  primaryButton: {
    backgroundColor: "#176df3",
    backgroundColorActive: "#235dba",
    backgroundColorHover: "#235dba",
    labelColor: "#f0f1f1",
    borderRadius: {
      type: "rounded",
      radius: "0.875em",
    },
  },
  secondaryButton: {
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
  phoneInputField: {
    borderRadius: {
      type: "rounded",
      radius: "0.875em",
    },
  },
  link: {
    color: "#2f7bf4",
  },
  logo: {
    height: "2.5rem",
  },
};

abstract class AbstractStyle<T> {
  abstract acceptDeclaration(declaration: Declaration): boolean;
  abstract acceptCssAstVisitor(visitor: CssAstVisitor): void;
  abstract getValue(): T;
  abstract setValue(value: T): void;
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
    this.setWithRawValue(declaration.value.trim());
    return true;
  }

  getValue(): T {
    return this.value;
  }

  setValue(value: T): void {
    this.value = value;
  }

  abstract getCSSValue(): string | undefined;
}

export class ColorStyleProperty extends StyleProperty<string | undefined> {
  protected setWithRawValue(rawValue: string): void {
    if (rawValue) {
      this.value = rawValue;
    } else {
      this.value = undefined;
    }
  }

  acceptCssAstVisitor(visitor: CssAstVisitor): void {
    visitor.visitColorStyleProperty(this);
  }

  getCSSValue(): string | undefined {
    return this.value;
  }
}

export class AlignItemsStyleProperty extends StyleProperty<
  Alignment | undefined
> {
  protected setWithRawValue(rawValue: string): void {
    switch (rawValue) {
      case "start":
        this.value = "start";
        break;
      case "end":
        this.value = "end";
        break;
      case "center":
        this.value = "center";
        break;
      default:
        this.value = undefined;
        break;
    }
  }

  acceptCssAstVisitor(visitor: CssAstVisitor): void {
    visitor.visitAlignItemsStyleProperty(this);
  }

  getCSSValue(): string | undefined {
    switch (this.value) {
      case "start":
        return "start";
      case "end":
        return "end";
      case "center":
        return "center";
      default:
        return undefined;
    }
  }
}

export class AlignOrHideItemsStyleProperty extends StyleProperty<
  Alignment | Hidden | undefined
> {
  protected setWithRawValue(rawValue: string): void {
    switch (rawValue) {
      case "start":
        this.value = "start";
        break;
      case "end":
        this.value = "end";
        break;
      case "center":
        this.value = "center";
        break;
      case "hidden":
        this.value = "hidden";
        break;
      default:
        this.value = undefined;
        break;
    }
  }

  acceptCssAstVisitor(visitor: CssAstVisitor): void {
    visitor.visitAlignOrHideItemsStyleProperty(this);
  }

  getCSSValue(): string | undefined {
    switch (this.value) {
      case "start":
        return "start";
      case "end":
        return "end";
      case "center":
        return "center";
      case "hidden":
        return "hidden";
      default:
        return undefined;
    }
  }
}

export class BorderRadiusStyleProperty extends StyleProperty<
  BorderRadiusStyle | undefined
> {
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
      case "":
        this.value = undefined;
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

  getCSSValue(): string | undefined {
    switch (this.value?.type) {
      case "rounded":
        return this.value.radius !== "" ? this.value.radius : undefined;
      case "rounded-full":
        return BorderRadiusStyleProperty.FULL_ROUNDED_CSS_VALUE;
      case "none":
        return "initial";
      default:
        return undefined;
    }
  }
}

export class SpaceStyleProperty extends StyleProperty<string | undefined> {
  protected setWithRawValue(rawValue: string): void {
    if (rawValue) {
      this.value = rawValue;
    } else {
      this.value = undefined;
    }
  }

  acceptCssAstVisitor(visitor: CssAstVisitor): void {
    visitor.visitSpaceStyleProperty(this);
  }

  getCSSValue(): string | undefined {
    return this.value;
  }
}

export class HeightStyleProperty extends StyleProperty<string | undefined> {
  protected setWithRawValue(rawValue: string): void {
    if (rawValue) {
      this.value = rawValue;
    } else {
      this.value = undefined;
    }
  }

  acceptCssAstVisitor(visitor: CssAstVisitor): void {
    visitor.visitHeightStyleProperty(this);
  }

  getCSSValue(): string | undefined {
    return this.value;
  }
}

type StyleProperties<T> = {
  [K in keyof T]: AbstractStyle<T[K] | null>;
};
export class StyleGroup<T extends object> extends AbstractStyle<T> {
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
      const s = style as AbstractStyle<unknown>;
      value[name] = s.getValue();
    }
    return value as T;
  }

  setValue(value: T): void {
    for (const [k, v] of Object.entries(value)) {
      const style = (this.styles as any)[k] as AbstractStyle<T>;
      style.setValue(v);
    }
  }
}

export class CustomisableThemeStyleGroup extends StyleGroup<PartialCustomisableTheme> {
  // eslint-disable-next-line complexity
  constructor(value: PartialCustomisableTheme = EMPTY_THEME) {
    super({
      page: new StyleGroup({
        backgroundColor: new ColorStyleProperty(
          CSSVariable.LayoutBackgroundColor,
          value.page.backgroundColor
        ),
      }),
      card: new StyleGroup({
        alignment: new AlignItemsStyleProperty(
          CSSVariable.AlignmentCard,
          value.card.alignment
        ),
      }),
      primaryButton: new StyleGroup({
        backgroundColor: new ColorStyleProperty(
          CSSVariable.PrimaryButtonBackgroundColor,
          value.primaryButton.backgroundColor
        ),
        backgroundColorActive: new ColorStyleProperty(
          CSSVariable.PrimaryButtonBackgroundColorActive,
          value.primaryButton.backgroundColorActive
        ),
        backgroundColorHover: new ColorStyleProperty(
          CSSVariable.PrimaryButtonBackgroundColorHover,
          value.primaryButton.backgroundColorHover
        ),
        labelColor: new ColorStyleProperty(
          CSSVariable.PrimaryButtonTextColor,
          value.primaryButton.labelColor
        ),
        borderRadius: new BorderRadiusStyleProperty(
          CSSVariable.PrimaryButtonBorderRadius,
          value.primaryButton.borderRadius
        ),
      }),

      secondaryButton: new StyleGroup({
        borderRadius: new BorderRadiusStyleProperty(
          CSSVariable.SecondaryButtonBorderRadius,
          value.secondaryButton.borderRadius
        ),
      }),

      inputField: new StyleGroup({
        borderRadius: new BorderRadiusStyleProperty(
          CSSVariable.InputFiledBorderRadius,
          value.inputField.borderRadius
        ),
      }),

      phoneInputField: new StyleGroup({
        borderRadius: new BorderRadiusStyleProperty(
          CSSVariable.PhoneInputTriggerBorderRadius,
          value.inputField.borderRadius
        ),
      }),

      link: new StyleGroup({
        color: new ColorStyleProperty(CSSVariable.LinkColor, value.link.color),
      }),

      logo: new StyleGroup({
        height: new HeightStyleProperty(
          CSSVariable.LogoHeight,
          value.logo.height
        ),
      }),
    });
  }
}

export class StyleCssVisitor<T extends object> extends CssNodeVisitor {
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

  visitStyleGroup<T extends object>(styleGroup: StyleGroup<T>): void {
    for (const style of Object.values(styleGroup.styles)) {
      const s = style as AbstractStyle<T>;
      s.acceptCssAstVisitor(this);
    }
  }

  visitAlignItemsStyleProperty(styleProperty: AlignItemsStyleProperty): void {
    this.visitorStyleProperty(styleProperty);
  }

  visitAlignOrHideItemsStyleProperty(
    styleProperty: AlignOrHideItemsStyleProperty
  ): void {
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

  visitSpaceStyleProperty(styleProperty: SpaceStyleProperty): void {
    this.visitorStyleProperty(styleProperty);
  }

  visitHeightStyleProperty(styleProperty: HeightStyleProperty): void {
    this.visitorStyleProperty(styleProperty);
  }

  visitorStyleProperty<T>(styleProperty: StyleProperty<T>): void {
    const value = styleProperty.getCSSValue();
    if (value == null || value === "") {
      return;
    }
    this.rule.append(
      new Declaration({
        prop: styleProperty.propertyName,
        value: value,
      })
    );
  }

  getCSS(): Root {
    return this.root;
  }

  getDeclarations(): Declaration[] {
    return this.rule.nodes.filter(
      (n): n is Declaration => n instanceof Declaration
    );
  }
}
