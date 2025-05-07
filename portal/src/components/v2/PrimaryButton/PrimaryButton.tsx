import React from "react";
import cn from "classnames";
import { Button } from "@radix-ui/themes";
import styles from "./PrimaryButton.module.css";

export interface PrimaryButtonProps {
  darkMode?: boolean;
  size: "1" | "2" | "3" | "4";
  highContrast?: boolean;
  disabled?: boolean;
  loading?: boolean;
  text?: React.ReactNode;

  type?: "button" | "reset" | "submit";
  onClick?: React.MouseEventHandler<HTMLButtonElement>;
}

export function PrimaryButton({
  darkMode,
  size,
  highContrast,
  disabled,
  loading,
  text,

  type = "button",
  onClick,
}: PrimaryButtonProps): React.ReactElement {
  return (
    <Button
      type={type}
      className={cn(styles.primaryButton, darkMode ? "dark" : null)}
      size={size}
      variant="solid"
      highContrast={highContrast}
      disabled={loading ? true : disabled}
      color="indigo"
      loading={loading}
      onClick={onClick}
    >
      {text}
    </Button>
  );
}
