import React from "react";
import cn from "classnames";
import styles from "./OnboardingSurveyLayout.module.css";
import { Logo } from "../common/Logo";
import backgroundImage from "../../images/onboarding-bg.svg";

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
      <div
        className={styles.onboardingSurveyLayout__bg}
        style={{ backgroundImage: `url(${backgroundImage})` }}
      >
        <div className={cn(styles.onboardingSurveyLayout__content, "dark")}>
          {children}
        </div>
      </div>
    </div>
  );
}
