import React from "react";
import styles from "./ProjectWizardLayout.module.css";
import ScreenHeader from "../../ScreenHeader";
import { ProjectWizardPreview } from "./ProjectWizardPreview";

function Header() {
  return <ScreenHeader showHamburger={false} />;
}

export interface ProjectWizardLayoutProps {
  children?: React.ReactNode;
}

export function ProjectWizardLayout({
  children,
}: ProjectWizardLayoutProps): React.ReactElement {
  return (
    <div className={styles.projectWizardLayout__root}>
      <Header />
      <div className={styles.projectWizardLayout__content}>
        <section className={styles.projectWizardLayout__left}>
          <div className={styles.projectWizardLayout__leftFormContainer}>
            {children}
          </div>
        </section>
        <section className={styles.projectWizardLayout__right}>
          <ProjectWizardPreview className="flex-1" />
        </section>
      </div>
    </div>
  );
}
