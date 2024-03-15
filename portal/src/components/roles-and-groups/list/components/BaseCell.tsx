import React, { PropsWithChildren } from "react";
import styles from "./BaseCell.module.css";

function BaseCell(props: PropsWithChildren<{}>): React.ReactElement {
  return <div className={styles.cell}>{props.children}</div>;
}

export default BaseCell;
