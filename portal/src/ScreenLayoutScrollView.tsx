import React from "react";
import styles from "./ScreenLayoutScrollView.module.css";

interface ScreenLayoutScrollViewProps {
  children: React.ReactNode;
}

const ScreenLayoutScrollView: React.FC<ScreenLayoutScrollViewProps> = (
  props: ScreenLayoutScrollViewProps
) => {
  const { children } = props;
  return <div className={styles.scrollView}>{children}</div>;
};

export default ScreenLayoutScrollView;
