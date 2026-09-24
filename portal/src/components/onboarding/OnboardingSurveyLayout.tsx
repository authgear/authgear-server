import React from "react";
import cn from "classnames";
import styles from "./OnboardingSurveyLayout.module.css";
import { Logo } from "../common/Logo";
import backgroundImage from "../../images/onboarding-bg.svg";

function Header() {
  return (
    <header className="block">
      {/* The artwork does not change with the appearance, so the mark must not
          either. */}
      <Logo variant="white" />
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
    <div
      className={styles.onboardingSurveyLayout__root}
      style={{ backgroundImage: `url(${backgroundImage})` }}
    >
      <Header />
      <div className={styles.onboardingSurveyLayout__bg}>
        {/* "dark" here is Radix's dark token set, not the appearance setting:
            the surface is the accent gradient in either appearance, so its
            content always needs light-on-dark tokens. */}
        <div className={cn(styles.onboardingSurveyLayout__content, "dark")}>
          {children}
        </div>
      </div>
    </div>
  );
}
