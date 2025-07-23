import React from "react";

import styles from "./ScreenContentHeader.module.css";

export interface ScreenContentHeaderProps {
  title?: React.ReactNode;
  description?: React.ReactNode;
  suffix?: React.ReactNode;
}

const ScreenContentHeader: React.VFC<ScreenContentHeaderProps> =
  function ScreenContentHeader(props: ScreenContentHeaderProps) {
    const { title, description, suffix } = props;
    return (
      <div className={styles.root}>
        <div className={styles.header}>
          {title}
          {description}
        </div>
        {suffix}
      </div>
    );
  };

export default ScreenContentHeader;
