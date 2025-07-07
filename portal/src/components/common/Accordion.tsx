import cn from "classnames";
import React, { useCallback, useState } from "react";
import { Text, FontIcon } from "@fluentui/react";
import styles from "./Accordion.module.css";
import ActionButton from "../../ActionButton";

export function Accordion({
  className,
  text,
  children,
}: {
  className?: string;
  text: React.ReactNode;
  children?: React.ReactNode;
}): React.ReactElement {
  const [isExpanded, setIsExpanded] = useState(false);

  const toggle = useCallback(() => {
    setIsExpanded((prev) => !prev);
  }, []);

  return (
    <div className={cn(className, styles.accordionRoot)}>
      <ActionButton
        className={styles.accordionToggle}
        type="button"
        onClick={toggle}
        styles={{
          root: {
            padding: 0,
            height: "auto",
          },
          label: {
            margin: 0,
          },
        }}
        text={
          <div className={styles.accordionToggleText}>
            <Text
              styles={{ root: { color: "inherit", lineHeight: "1.25rem" } }}
              variant="medium"
            >
              {text}
            </Text>
            <FontIcon
              className="w-4 h-4 text-base leading-none ml-2"
              iconName={isExpanded ? "ChevronUp" : "ChevronDown"}
            />
          </div>
        }
      />
      <div
        className={cn(
          styles.accordionContent,
          isExpanded ? null : styles["accordionContent--hide"]
        )}
      >
        {children}
      </div>
    </div>
  );
}
