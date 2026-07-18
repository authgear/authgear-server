import React from "react";
import { Text } from "@fluentui/react";
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
        <Text
          variant="mediumPlus"
          styles={{
            root: {
              fontWeight: 600,
              color: "var(--gray-12)",
              backgroundColor: "var(--gray-a3)",
              width: 22,
              height: 22,
              borderRadius: 999,
              textAlign: "center",
              lineHeight: 20,
            },
          }}
          block={true}
        >
          {stepNumber}
        </Text>
        <Text
          variant="mediumPlus"
          styles={{
            root: {
              fontWeight: 600,
              color: "var(--gray-12)",
            },
          }}
        >
          {title}
        </Text>
      </header>
      <div className={styles.quickStartStep__childrenContainer}>{children}</div>
    </section>
  );
}
