import React, { PropsWithChildren } from "react";
import BaseCell from "./BaseCell";
import styles from "./TextCell.module.css";
import { Text } from "@fluentui/react";

function TextCell(
  props: PropsWithChildren<Record<never, never>>
): React.ReactElement {
  return (
    <BaseCell>
      <Text className={styles.cellText}>{props.children}</Text>
    </BaseCell>
  );
}

export default TextCell;
