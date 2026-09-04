import React from "react";
import { Text } from "@radix-ui/themes";
import styles from "./QuickStartStep.module.css";

export function QuickStartStep({
  className,
  stepNumber,
  title,
  children,
}: {
  className?: string;
  stepNumber: string;
  title: React.ReactNode;
  children: React.ReactNode;
}): React.ReactElement {
  return (
    <section className={className}>
      <header className={styles.quickStartStep__header}>
        <Text size="3" weight="bold" className={styles.quickStartStep__number}>
          {stepNumber}
        </Text>
        <Text size="3" weight="bold" className={styles.quickStartStep__title}>
          {title}
        </Text>
      </header>
      <div className={styles.quickStartStep__childrenContainer}>{children}</div>
    </section>
  );
}
