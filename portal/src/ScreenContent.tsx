import React from "react";
import styles from "./ScreenContent.module.scss";

export interface ScreenContentProps {
  children?: React.ReactNode;
}

const ScreenContent: React.FC<ScreenContentProps> = function ScreenContent(
  props: ScreenContentProps
) {
  const { children } = props;
  return <div className={styles.root}>{children}</div>;
};

export default ScreenContent;
