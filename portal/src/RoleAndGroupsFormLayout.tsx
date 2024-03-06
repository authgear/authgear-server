import React from "react";
import styles from "./RoleAndGroupsFormLayout.module.css";
import NavBreadcrumb, { BreadcrumbItem } from "./NavBreadcrumb";
import { useFormContainerBaseContext } from "./FormContainerBase";
import { FormErrorMessageBar } from "./FormErrorMessageBar";
import { ProgressIndicator } from "@fluentui/react";

interface RoleAndGroupsFormLayoutProps {
  breadcrumbs: BreadcrumbItem[];
  Footer?: React.ReactNode;
}

const progressIndicatorStyles = {
  itemProgress: {
    padding: 0,
  },
};

export const RoleAndGroupsFormLayout: React.VFC<
  React.PropsWithChildren<RoleAndGroupsFormLayoutProps>
> = function RoleAndGroupsFormLayout({ breadcrumbs, children, Footer }) {
  const { onSubmit, isUpdating } = useFormContainerBaseContext();

  return (
    <div className={styles.root}>
      <ProgressIndicator
        styles={progressIndicatorStyles}
        className={!isUpdating ? "hidden" : ""}
        barHeight={4}
      />
      <FormErrorMessageBar />
      <form onSubmit={onSubmit} noValidate={true} className={styles.main}>
        <NavBreadcrumb className={styles.header} items={breadcrumbs} />
        <section className={styles.content}>{children}</section>
        {Footer != null ? (
          <footer className={styles.footer}>{Footer}</footer>
        ) : null}
      </form>
    </div>
  );
};
