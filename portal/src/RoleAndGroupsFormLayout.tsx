import React from "react";
import styles from "./RoleAndGroupsFormLayout.module.css";
import NavBreadcrumb, { BreadcrumbItem } from "./NavBreadcrumb";
import { useFormContainerBaseContext } from "./FormContainerBase";
import { FormErrorMessageBar } from "./FormErrorMessageBar";

interface RoleAndGroupsFormLayoutProps {
  breadcrumbs: BreadcrumbItem[];
  Footer?: React.ReactNode;
}

export const RoleAndGroupsFormLayout: React.VFC<
  React.PropsWithChildren<RoleAndGroupsFormLayoutProps>
> = function RoleAndGroupsFormLayout({ breadcrumbs, children, Footer }) {
  const { onSubmit } = useFormContainerBaseContext();

  return (
    <div className={styles.root}>
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
