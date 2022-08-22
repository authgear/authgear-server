import React from "react";
import ScreenHeader from "./ScreenHeader";
import styles from "./WizardScreenLayout.module.css";

export interface WizardScreenLayoutProps {
  children?: React.ReactNode;
}

export default function WizardScreenLayout(
  props: WizardScreenLayoutProps
): React.ReactElement {
  const { children } = props;
  return (
    <div className={styles.root}>
      <ScreenHeader showHamburger={false} />
      {children}
    </div>
  );
}
