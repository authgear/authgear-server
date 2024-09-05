import React from "react";
import cn from "classnames";
import styles from "./ScreenContent.module.css";

export interface ScreenContentProps {
  className?: string;
  layout?: "list" | "auto-rows";
  header?: React.ReactNode;
  children?: React.ReactNode;
}

const ScreenContent: React.VFC<ScreenContentProps> = function ScreenContent(
  props: ScreenContentProps
) {
  const { className, header, children, layout = "auto-rows" } = props;
  return (
    <>
      {header != null ? <div className={styles.container}>{header}</div> : null}
      <div
        className={cn(
          className,
          styles.root,
          styles.container,
          layout === "list" ? styles.list : styles.autoRows
        )}
      >
        {children}
      </div>
    </>
  );
};

export default ScreenContent;
