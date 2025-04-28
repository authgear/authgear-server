import React from "react";
import cn from "classnames";
import { Button } from "@radix-ui/themes";
import styles from "./WhiteButton.module.css";

export interface WhiteButtonProps {
  size: "1" | "2" | "3" | "4";
  disabled?: boolean;
  loading?: boolean;
  text?: React.ReactNode;

  type?: "button" | "reset" | "submit";
  onClick?: React.MouseEventHandler<HTMLButtonElement>;
}

export function WhiteButton({
  size,
  disabled,
  loading,
  text,

  type = "button",
  onClick,
}: WhiteButtonProps): React.ReactElement {
  return (
    <Button
      type={type}
      // Only dark mode is supported
      // We need radix-themes here so the focus colors (e.g., --focus-8) is correct
      className={cn(styles.whiteButton, "radix-themes", "dark")}
      size={size}
      variant="solid"
      highContrast={false}
      disabled={disabled}
      loading={loading}
      onClick={onClick}
    >
      {text}
    </Button>
  );
}
