import React from "react";
import cn from "classnames";
import { Button } from "@radix-ui/themes";
import { ArrowLeftIcon } from "@radix-ui/react-icons";
import styles from "./TextButton.module.css";

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

export type TextButtonSize = "3" | "4";

export enum TextButtonIcon {
  Back = "Back",
}

function sizeToIconDimension(size: TextButtonSize) {
  switch (size) {
    case "3":
      return 18;
    case "4":
      return 20;
  }
}

function Icon({
  icon,
  size,
}: {
  icon: TextButtonIcon;
  size: TextButtonSize;
}): React.ReactElement {
  const dimension = sizeToIconDimension(size);
  switch (icon) {
    // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
    case TextButtonIcon.Back:
      return <ArrowLeftIcon width={dimension} height={dimension} />;
  }
}

export interface TextButtonProps {
  variant: TextButtonVariant;
  size: TextButtonSize;
  darkMode?: boolean;
  disabled?: boolean;
  loading?: boolean;
  text?: React.ReactNode;
  iconStart?: TextButtonIcon;

  type?: "button" | "reset" | "submit";
  onClick?: React.MouseEventHandler<HTMLButtonElement>;
}

export function TextButton({
  variant,
  size,
  darkMode,
  disabled,
  loading,
  text,
  iconStart,

  type = "button",
  onClick,
}: TextButtonProps): React.ReactElement {
  return (
    <Button
      type={type}
      className={cn(styles.textButton, darkMode ? "dark" : null)}
      size={size}
      variant="ghost"
      highContrast={variantToHighContrast(variant)}
      disabled={disabled}
      color={variantToColor(variant)}
      loading={loading}
      onClick={onClick}
    >
      {iconStart != null ? <Icon icon={iconStart} size={size} /> : null}
      {text}
    </Button>
  );
}
