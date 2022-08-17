import React from "react";
import cn from "classnames";
import styles from "./ScreenContent.module.css";

export interface ScreenContentProps {
  className?: string;
  layout?: "list" | "auto-rows";
  children?: React.ReactNode;
}

const ScreenContent: React.FC<ScreenContentProps> = function ScreenContent(
  props: ScreenContentProps
) {
  const { className, children, layout = "auto-rows" } = props;
  return (
    <div
      className={cn(
        className,
        styles.root,
        layout === "list" ? styles.list : "auto-rows"
      )}
    >
      {children}
    </div>
  );
};

export default ScreenContent;
