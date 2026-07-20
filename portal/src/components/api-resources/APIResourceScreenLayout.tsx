import React, { useMemo } from "react";
import { Text } from "@radix-ui/themes";
import { ChevronLeftIcon } from "@radix-ui/react-icons";
import { useParams } from "react-router-dom";
import { BreadcrumbItem } from "../../NavBreadcrumb";
import {
  ErrorMessageBar,
  ErrorMessageBarContextProvider,
} from "../../ErrorMessageBar";
import ScreenContent from "../../ScreenContent";
import ScreenLayoutScrollView from "../../ScreenLayoutScrollView";
import Link from "../../Link";
import styles from "./APIResourceScreenLayout.module.css";

export interface APIResourceLayoutProps {
  children?: React.ReactNode;
  breadcrumbItems: BreadcrumbItem[];
  headerDescription?: React.ReactNode;
  headerSuffix?: React.ReactNode;
  /** Defaults to `"list"` (full-width). Use `"auto-rows"` for constrained detail pages. */
  layout?: "list" | "auto-rows";
}

function resolveBreadcrumbPath(to: string, appID: string): string {
  if (to === "") {
    return "";
  }
  return to.replace("~/", `/project/${appID}/`);
}

const APIResourceScreenLayout: React.VFC<APIResourceLayoutProps> =
  function APIResourceScreenLayout({
    children,
    breadcrumbItems,
    headerDescription,
    headerSuffix,
    layout = "list",
  }) {
    const { appID } = useParams() as { appID: string };

    const { title, backLink } = useMemo(() => {
      const titleItem = breadcrumbItems[breadcrumbItems.length - 1];
      const parentItem =
        breadcrumbItems.length > 1
          ? breadcrumbItems[breadcrumbItems.length - 2]
          : null;
      const parentTo =
        parentItem != null
          ? resolveBreadcrumbPath(parentItem.to, appID)
          : "";

      return {
        title: titleItem?.label ?? null,
        backLink:
          parentItem != null && parentTo !== ""
            ? { to: parentTo, label: parentItem.label }
            : null,
      };
    }, [appID, breadcrumbItems]);

    return (
      <ErrorMessageBarContextProvider>
        <ScreenLayoutScrollView>
          <div className={styles.scrollBody}>
            <ErrorMessageBar />
            <ScreenContent layout={layout} className={styles.screenContent}>
              <div
                className={
                  layout === "auto-rows"
                    ? styles.headerConstrained
                    : styles.header
                }
              >
                <div className={styles.headerMain}>
                  {backLink != null ? (
                    <Link to={backLink.to} className={styles.backLink}>
                      <ChevronLeftIcon className={styles.backLinkIcon} />
                      <span>{backLink.label}</span>
                    </Link>
                  ) : null}
                  <div className={styles.titleRow}>
                    <Text
                      as="p"
                      size="5"
                      weight="bold"
                      className={styles.pageTitle}
                    >
                      {title}
                    </Text>
                    {headerSuffix}
                  </div>
                  {headerDescription}
                </div>
              </div>
              {children}
            </ScreenContent>
          </div>
        </ScreenLayoutScrollView>
      </ErrorMessageBarContextProvider>
    );
  };

export default APIResourceScreenLayout;
