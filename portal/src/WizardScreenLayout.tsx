import React from "react";
import ScreenHeader from "./ScreenHeader";
import styles from "./WizardScreenLayout.module.scss";

export interface WizardScreenLayoutProps {
  children?: React.ReactNode;
}

export default function WizardScreenLayout(
  props: WizardScreenLayoutProps
): React.ReactElement {
  const { children } = props;
  return (
    <div className={styles.root}>
      <ScreenHeader />
      {children}
    </div>
  );
}
