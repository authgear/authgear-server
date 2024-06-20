type Color = string;

export type BorderRadiusStyleType = "none" | "rounded" | "rounded-full";

export type BorderRadiusStyle =
  | {
      type: "none";
    }
  | {
      type: "rounded";
      radius: number;
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

export interface CustomisableTheme {
  cardAlignment: "left" | "center" | "right";
  backgroundColor: Color;

  primaryButton: ButtonStyle;
  inputField: InputFieldStyle;

  linkColor: Color;
}
