import React, { PropsWithChildren } from "react";
import BaseCell from "./BaseCell";
import styles from "./TextCell.module.css";

function TextCell(props: PropsWithChildren<{}>) {
  return (
    <BaseCell>
      <div className={styles.cellText}>{props.children}</div>
    </BaseCell>
  );
}

export default TextCell;
