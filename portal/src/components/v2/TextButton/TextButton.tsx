import React from "react";
import cn from "classnames";
import { Button } from "@radix-ui/themes";

export type TextButtonVariant = "default" | "secondary";

function variantToColor(type: TextButtonVariant) {
  switch (type) {
    case "default":
      return "indigo";
    case "secondary":
      return "gray";
  }
}

function variantToHighContrast(type: TextButtonVariant): boolean {
  switch (type) {
    case "default":
      return false;
    case "secondary":
      return true;
  }
}

export interface TextButtonProps {
  variant: TextButtonVariant;
  size: "1" | "2" | "3" | "4";
  darkMode?: boolean;
  disabled?: boolean;
  loading?: boolean;
  text?: React.ReactNode;
  iconStart?: React.ReactNode;
}

export function TextButton({
  variant,
  size,
  darkMode,
  disabled,
  loading,
  text,
  iconStart,
}: TextButtonProps): React.ReactElement {
  return (
    <Button
      className={cn(darkMode ? "dark" : null)}
      size={size}
      variant="ghost"
      highContrast={variantToHighContrast(variant)}
      disabled={disabled}
      color={variantToColor(variant)}
      loading={loading}
    >
      {iconStart}
      {text}
    </Button>
  );
}
