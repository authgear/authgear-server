import React from "react";
import cn from "classnames";
import styles from "./ScreenContent.module.scss";

export interface ScreenContentProps {
  className?: string;
  children?: React.ReactNode;
}

const ScreenContent: React.FC<ScreenContentProps> = function ScreenContent(
  props: ScreenContentProps
) {
  const { className, children } = props;
  return <div className={cn(styles.root, className)}>{children}</div>;
};

export default ScreenContent;
