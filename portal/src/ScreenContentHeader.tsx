import React from "react";

import styles from "./ScreenContentHeader.module.css";

export interface ScreenContentHeaderProps {
  title?: React.ReactNode;
  description?: React.ReactNode;
}

const ScreenContentHeader: React.VFC<ScreenContentHeaderProps> =
  function ScreenContentHeader(props: ScreenContentHeaderProps) {
    const { title, description } = props;
    return (
      <div className={styles.root}>
        {title}
        {description}
      </div>
    );
  };

export default ScreenContentHeader;
