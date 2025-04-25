import { Tooltip as RadixTooltip } from "@radix-ui/themes";
import React from "react";

export interface TooltipProps {
  content: React.ReactNode;
  disabled?: boolean;

  children?: React.ReactNode;
}

export function Tooltip({
  content,
  disabled,
  children,
}: TooltipProps): React.ReactElement {
  return (
    <RadixTooltip content={content} open={disabled ? false : undefined}>
      {children}
    </RadixTooltip>
  );
}
