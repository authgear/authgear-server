type Color = string;

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
  cardAlignment: "left" | "center" | "right";
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
  abstract accept(declaration: Declaration): boolean;
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

  accept(declaration: Declaration): boolean {
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

  accept(declaration: Declaration): boolean {
    for (const style of Object.values(this.styles)) {
      const s = style as AbstractStyle<T>;
      if (s.accept(declaration)) {
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
