import React from "react";
import { Badge as RadixBadge } from "@radix-ui/themes";
import { semanticToRadixColor } from "../../../util/radix";

type BadgeSize = "1" | "2";
type BadgeVariant = "info" | "neutral" | "success" | "warning" | "error";

export interface BadgeProps {
  size: BadgeSize;
  variant: BadgeVariant;
  className?: string;
  text?: React.ReactNode;
}

export function Badge({
  text,
  size,
  variant,
  className,
}: BadgeProps): React.ReactElement {
  return (
    <RadixBadge
      size={size}
      variant="soft"
      color={variantToRadixColor(variant)}
      className={className}
    >
      {text}
    </RadixBadge>
  );
}

function variantToRadixColor(
  variant: BadgeVariant
): React.ComponentProps<typeof RadixBadge>["color"] {
  switch (variant) {
    case "info":
      return "indigo";
    case "neutral":
      return "gray";
    case "success":
      return semanticToRadixColor("success");
    case "warning":
      return semanticToRadixColor("warning");
    case "error":
      return semanticToRadixColor("error");
  }
}
