import React, { useMemo } from "react";
import cn from "classnames";

import styles from "./DefaultLayout.module.css";

interface DefaultLayoutHeaderComponentProps {
  children?: React.ReactNode;
}

export interface DefaultLayoutProps {
  position: "sticky" | "end";
  className?: string;
  messageBar?: React.ReactNode;
  footer?: React.ReactNode;
  HeaderComponent?: React.VFC<DefaultLayoutHeaderComponentProps>;
  children?: React.ReactNode;
}

const DefaultLayout: React.VFC<DefaultLayoutProps> = function DefaultLayout(
  props
) {
  const { position, className, messageBar, footer, HeaderComponent, children } =
    props;

  const defaultHeaderContent = useMemo(() => messageBar, [messageBar]);

  return (
    <div className={cn(styles.container, className)}>
      <div className={styles.header}>
        {HeaderComponent != null ? (
          <HeaderComponent>{defaultHeaderContent}</HeaderComponent>
        ) : (
          defaultHeaderContent
        )}
      </div>
      <div className={cn(position === "sticky" && styles.contentExpand)}>
        {children}
      </div>
      {footer != null ? (
        <div
          className={cn(
            styles.footer,
            position === "sticky" && styles.footerSticky
          )}
        >
          {footer}
        </div>
      ) : null}
    </div>
  );
};

export default DefaultLayout;
