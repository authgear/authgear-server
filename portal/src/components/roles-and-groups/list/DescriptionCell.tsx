import React, { PropsWithChildren } from "react";
import BaseCell from "./BaseCell";
import styles from "./DescriptionCell.module.css";

function DescriptionCell(props: PropsWithChildren<{}>) {
  return (
    <BaseCell>
      <div className={styles.description}>{props.children}</div>
    </BaseCell>
  );
}

export default DescriptionCell;
