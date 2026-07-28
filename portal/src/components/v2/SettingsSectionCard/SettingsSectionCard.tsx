import React from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import styles from "./SettingsSectionCard.module.css";

export interface SettingsSectionCardProps {
  /** The label shown on the left (top, when narrow). */
  title: React.ReactNode;
  /** Optional supporting text shown under the title. */
  description?: React.ReactNode;
  /**
   * `columns` (default): title on the left, content on the right (stacks on
   * tablet). `stacked`: title above content in a single column.
   */
  layout?: "columns" | "stacked";
  /** Extra classes for the outer card (e.g. grid placement, save-bar clearance). */
  className?: string;
  /** Extra classes for the content column (e.g. the gap between fields). */
  contentClassName?: string;
  children: React.ReactNode;
}

/**
 * A bordered settings card. Default layout is a label column on the left and a
 * content column on the right (stacks on narrow viewports). Use `layout="stacked"`
 * to keep title and content in one vertical column.
 */
export function SettingsSectionCard({
  title,
  description,
  layout = "columns",
  className,
  contentClassName,
  children,
}: SettingsSectionCardProps): React.ReactElement {
  return (
    <div
      className={cn(
        styles.card,
        layout === "stacked" && styles["card--stacked"],
        className
      )}
    >
      <div
        className={cn(
          styles.titleColumn,
          layout === "stacked" && styles["titleColumn--stacked"]
        )}
      >
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
