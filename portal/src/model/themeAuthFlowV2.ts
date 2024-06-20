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
