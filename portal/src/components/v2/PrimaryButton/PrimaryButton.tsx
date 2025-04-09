import React from "react";
import cn from "classnames";
import { Button } from "@radix-ui/themes";

export interface PrimaryButtonProps {
  darkMode?: boolean;
  size: "1" | "2" | "3" | "4";
  highContrast?: boolean;
  disabled?: boolean;
  loading?: boolean;
  text?: React.ReactNode;

  onClick?: React.MouseEventHandler<HTMLButtonElement>;
}

export function PrimaryButton({
  darkMode,
  size,
  highContrast,
  disabled,
  loading,
  text,
  onClick,
}: PrimaryButtonProps): React.ReactElement {
  return (
    <Button
      className={cn(darkMode ? "dark" : null)}
      size={size}
      variant="solid"
      highContrast={highContrast}
      disabled={disabled}
      color="indigo"
      loading={loading}
      onClick={onClick}
    >
      {text}
    </Button>
  );
}
