import React from "react";
import cn from "classnames";
import styles from "./RoleAndGroupsLayout.module.css";
import NavBreadcrumb, { BreadcrumbItem } from "./NavBreadcrumb";
import { ProgressIndicator } from "@fluentui/react";
import { useIsLoading } from "./hook/loading";
import {
  ErrorMessageBar,
  ErrorMessageBarContextProvider,
} from "./ErrorMessageBar";

interface RoleAndGroupsLayoutProps {
  breadcrumbs: BreadcrumbItem[];
}

const progressIndicatorStyles = {
  itemProgress: {
    padding: 0,
  },
};

export const RoleAndGroupsLayout: React.VFC<
  React.PropsWithChildren<RoleAndGroupsLayoutProps>
> = function RoleAndGroupsLayout({ breadcrumbs, children }) {
  const isLoading = useIsLoading();

  return (
    <ErrorMessageBarContextProvider>
      <div className={styles.root}>
        <div className={styles.topBar}>
          <ProgressIndicator
            styles={progressIndicatorStyles}
            className={!isLoading ? "hidden" : ""}
            barHeight={4}
          />
          <ErrorMessageBar />
        </div>
        <div className={styles.main}>
          <NavBreadcrumb className={styles.header} items={breadcrumbs} />
          <section className={styles.content}>{children}</section>
        </div>
      </div>
    </ErrorMessageBarContextProvider>
  );
};

export const RoleAndGroupsVeriticalFormLayout: React.VFC<
  React.PropsWithChildren<Record<never, never>>
> = function RoleAndGroupsVeriticalFormLayout({ children }) {
  return <div className={styles.verticalForm}>{children}</div>;
};

export const RoleAndGroupsFormFooter: React.VFC<
  React.PropsWithChildren<{ className?: string }>
> = function RoleAndGroupsFormFooter({ children, className }) {
  return (
    <footer className={cn(styles.formFooter, className)}>{children}</footer>
  );
};
