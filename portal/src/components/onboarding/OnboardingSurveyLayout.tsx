import React from "react";
import styles from "./OnboardingSurveyLayout.module.css";
import { Logo } from "../common/Logo";

function Header() {
  return (
    <header className="block">
      <Logo inverted={true} />
    </header>
  );
}

export interface OnboardingSurveyLayoutProps {
  children?: React.ReactNode;
}

export function OnboardingSurveyLayout({
  children,
}: OnboardingSurveyLayoutProps): React.ReactElement {
  return (
    <div className={styles.onboardingSurveyLayout__root}>
      <Header />
      <div className={styles.onboardingSurveyLayout__content}>{children}</div>
    </div>
  );
}
