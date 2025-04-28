import React from "react";
import cn from "classnames";
import { Button } from "@radix-ui/themes";
import styles from "./SecondaryButton.module.css";

export interface SecondaryButtonProps {
  size: "1" | "2" | "3" | "4";
  disabled?: boolean;
  loading?: boolean;
  text?: React.ReactNode;

  type?: "button" | "reset" | "submit";
  onClick?: React.MouseEventHandler<HTMLButtonElement>;
}

export function SecondaryButton({
  size,
  disabled,
  loading,
  text,

  type = "button",
  onClick,
}: SecondaryButtonProps): React.ReactElement {
  return (
    <Button
      type={type}
      className={cn(styles.secondaryButton)}
      size={size}
      variant="outline"
      highContrast={false}
      disabled={disabled}
      color="gray"
      loading={loading}
      onClick={onClick}
    >
      {text}
    </Button>
  );
}
