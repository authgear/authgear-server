import React from "react";
import cn from "classnames";
import { Button } from "@radix-ui/themes";

export type TextButtonVariant = "default" | "secondary";

function varientToColor(type: TextButtonVariant) {
  switch (type) {
    case "default":
      return "indigo";
    case "secondary":
      return "gray";
  }
}

function varientToHighContrast(type: TextButtonVariant): boolean {
  switch (type) {
    case "default":
      return false;
    case "secondary":
      return true;
  }
}

export interface TextButtonProps {
  varient: TextButtonVariant;
  size: "1" | "2" | "3" | "4";
  darkMode?: boolean;
  disabled?: boolean;
  text?: React.ReactNode;
  iconStart?: React.ReactNode;
}

export function TextButton({
  varient: variant,
  size,
  darkMode,
  disabled,
  text,
  iconStart,
}: TextButtonProps): React.ReactElement {
  return (
    <Button
      className={cn(darkMode ? "dark" : null)}
      size={size}
      variant="ghost"
      highContrast={varientToHighContrast(variant)}
      disabled={disabled}
      color={varientToColor(variant)}
    >
      {iconStart}
      {text}
    </Button>
  );
}
