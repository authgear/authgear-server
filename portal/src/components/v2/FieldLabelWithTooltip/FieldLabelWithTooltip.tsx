import React from "react";
import cn from "classnames";
import { InfoCircledIcon } from "@radix-ui/react-icons";
import { Tooltip } from "../Tooltip/Tooltip";
import styles from "./FieldLabelWithTooltip.module.css";

export interface FieldLabelWithTooltipProps {
  className?: string;
  /** The label text, plus any marks that should sit before the icon. */
  children: React.ReactNode;
  /** Shown when the info icon is hovered. */
  tooltip: React.ReactNode;
  /** Plain-text tooltip for assistive technology. */
  tooltipLabel: string;
}

// A field label followed by an info icon that reveals a tooltip. Pass it as
// the `label` of FormField / TextField.
export function FieldLabelWithTooltip({
  className,
  children,
  tooltip,
  tooltipLabel,
}: FieldLabelWithTooltipProps): React.ReactElement {
  return (
    <span className={cn(styles.root, className)}>
      {children}
      <Tooltip content={tooltip}>
        <InfoCircledIcon
          className={styles.icon}
          role="img"
          aria-label={tooltipLabel}
        />
      </Tooltip>
    </span>
  );
}
