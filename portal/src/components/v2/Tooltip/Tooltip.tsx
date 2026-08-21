import { Tooltip as RadixTooltip } from "@radix-ui/themes";
import React from "react";
import styles from "./Tooltip.module.css";

export interface TooltipProps {
  content: React.ReactNode;
  disabled?: boolean;
  side?: "top" | "right" | "bottom" | "left";
  children?: React.ReactNode;
}

export function Tooltip({
  content,
  disabled,
  side = "top",
  children,
}: TooltipProps): React.ReactElement {
  return (
    <RadixTooltip
      content={content}
      open={disabled ? false : undefined}
      side={side}
      className={styles.tooltip}
    >
      {children}
    </RadixTooltip>
  );
}
