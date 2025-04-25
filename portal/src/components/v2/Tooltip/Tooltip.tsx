import { Tooltip as RadixTooltip } from "@radix-ui/themes";
import React from "react";

export interface TooltipProps {
  content: React.ReactNode;

  children?: React.ReactNode;
}

export function Tooltip({
  content,
  children,
}: TooltipProps): React.ReactElement {
  return <RadixTooltip content={content}>{children}</RadixTooltip>;
}
