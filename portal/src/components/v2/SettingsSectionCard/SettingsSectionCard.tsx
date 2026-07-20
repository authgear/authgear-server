import React from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import styles from "./SettingsSectionCard.module.css";

export interface SettingsSectionCardProps {
  /** The label shown on the left (top, when narrow). */
  title: React.ReactNode;
  /** Optional supporting text shown under the title. */
  description?: React.ReactNode;
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
  description,
  className,
  contentClassName,
  children,
}: SettingsSectionCardProps): React.ReactElement {
  return (
    <div className={cn(styles.card, className)}>
      <div className={styles.titleColumn}>
        <Text as="p" size="3" weight="medium" className={styles.title}>
          {title}
        </Text>
        {description != null ? (
          <Text as="p" size="2" color="gray" className={styles.description}>
            {description}
          </Text>
        ) : null}
      </div>
      <div className={cn(styles.content, contentClassName)}>{children}</div>
    </div>
  );
}
