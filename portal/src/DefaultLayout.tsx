import React from "react";
import cn from "classnames";

import styles from "./DefaultLayout.module.css";

export interface DefaultLayoutProps {
  className?: string;
  messageBar?: React.ReactNode;
  footer?: React.ReactNode;
  footerPosition?: "sticky" | "end";
  children?: React.ReactNode;
}

const DefaultLayout: React.VFC<DefaultLayoutProps> = function DefaultLayout(
  props
) {
  const { className, messageBar, footer, footerPosition, children } = props;
  return (
    <div className={cn(styles.container, className)}>
      <div className={styles.header}>{messageBar}</div>
      <div className={cn(footerPosition === "sticky" && styles.contentExpand)}>
        {children}
      </div>
      {footer != null ? (
        <div
          className={cn(
            styles.footer,
            footerPosition === "sticky" && styles.footerSticky
          )}
        >
          {footer}
        </div>
      ) : null}
    </div>
  );
};

export default DefaultLayout;
