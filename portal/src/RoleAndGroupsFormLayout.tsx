import React from "react";
import styles from "./RoleAndGroupsFormLayout.module.css";
import NavBreadcrumb, { BreadcrumbItem } from "./NavBreadcrumb";
import { useFormContainerBaseContext } from "./FormContainerBase";

interface RoleAndGroupsFormLayoutProps {
  breadcrumbs: BreadcrumbItem[];
  Footer?: React.ReactNode;
}

export const RoleAndGroupsFormLayout: React.VFC<
  React.PropsWithChildren<RoleAndGroupsFormLayoutProps>
> = function RoleAndGroupsFormLayout({ breadcrumbs, children, Footer }) {
  const { onSubmit } = useFormContainerBaseContext();

  return (
    <form onSubmit={onSubmit} className={styles.root}>
      <NavBreadcrumb className={styles.header} items={breadcrumbs} />
      <section className={styles.content}>{children}</section>
      {Footer != null ? (
        <footer className={styles.footer}>{Footer}</footer>
      ) : null}
    </form>
  );
};
