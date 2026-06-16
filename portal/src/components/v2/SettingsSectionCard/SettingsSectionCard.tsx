import React from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import styles from "./SettingsSectionCard.module.css";

export interface SettingsSectionCardProps {
  /** The label shown on the left (top, when narrow). */
  title: React.ReactNode;
  /** Extra classes for the outer card (e.g. grid placement, save-bar clearance). */
  className?: string;
  /** Extra classes for the content column (e.g. the gap between fields). */
  contentClassName?: string;
  children: React.ReactNode;
}

/**
 * A bordered settings card laid out as a label column on the left and a content
 * column on the right, stacking vertically on narrow (tablet) viewports.
 */
export function SettingsSectionCard({
  title,
  className,
  contentClassName,
  children,
}: SettingsSectionCardProps): React.ReactElement {
  return (
    <div className={cn(styles.card, className)}>
      <Text as="p" size="3" weight="medium" className={styles.title}>
        {title}
      </Text>
      <div className={cn(styles.content, contentClassName)}>{children}</div>
    </div>
  );
}
