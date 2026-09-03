import React, { ReactNode, useMemo } from "react";
import cn from "classnames";
import { Heading } from "@radix-ui/themes";
import { ChevronLeftIcon } from "@radix-ui/react-icons";
import { useParams } from "react-router-dom";
import styles from "./RoleAndGroupsLayout.module.css";
import { BreadcrumbItem } from "./NavBreadcrumb";
import {
  ErrorMessageBar,
  ErrorMessageBarContextProvider,
} from "./ErrorMessageBar";
import Link from "./Link";

interface RoleAndGroupsLayoutProps {
  headerBreadcrumbs: BreadcrumbItem[];
  headerSubitem?: ReactNode;
  headerDescription?: ReactNode;
}

function resolveBreadcrumbPath(to: string, appID: string): string {
  if (to === "" || to === ".") {
    return "";
  }
  return to.replace("~/", `/project/${appID}/`);
}

export const RoleAndGroupsLayout: React.VFC<
  React.PropsWithChildren<RoleAndGroupsLayoutProps>
> = function RoleAndGroupsLayout({
  headerBreadcrumbs,
  headerSubitem,
  headerDescription,
  children,
}) {
  const { appID } = useParams() as { appID: string };

  const { title, backLink } = useMemo(() => {
    const titleItem = headerBreadcrumbs[headerBreadcrumbs.length - 1];
    const parentItem =
      headerBreadcrumbs.length > 1
        ? headerBreadcrumbs[headerBreadcrumbs.length - 2]
        : null;
    const parentTo =
      parentItem != null ? resolveBreadcrumbPath(parentItem.to, appID) : "";

    return {
      title: titleItem.label ?? null,
      backLink:
        parentItem != null && parentTo !== ""
          ? { to: parentTo, label: parentItem.label }
          : null,
    };
  }, [appID, headerBreadcrumbs]);

  return (
    <ErrorMessageBarContextProvider>
      <div className={styles.root}>
        <div className={styles.topBar}>
          <ErrorMessageBar />
        </div>
        <div className={styles.main}>
          <header className={styles.header}>
            <div className={styles.headerMain}>
              {backLink != null ? (
                <Link to={backLink.to} className={styles.backLink}>
                  <ChevronLeftIcon className={styles.backLinkIcon} />
                  <span>{backLink.label}</span>
                </Link>
              ) : null}
              <div className={styles.titleRow}>
                <Heading
                  as="h1"
                  size="5"
                  weight="bold"
                  className={styles.pageTitle}
                >
                  {title}
                </Heading>
                {headerSubitem != null ? headerSubitem : null}
              </div>
              {headerDescription != null ? (
                <div className={styles.headerDescription}>
                  {headerDescription}
                </div>
              ) : null}
            </div>
          </header>
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
