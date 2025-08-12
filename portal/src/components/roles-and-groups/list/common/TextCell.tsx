import React, { PropsWithChildren } from "react";
import BaseCell from "./BaseCell";
import styles from "./TextCell.module.css";
import { Text } from "@fluentui/react";

function TextCell(
  props: PropsWithChildren<Record<never, never>>
): React.ReactElement {
  return (
    <BaseCell>
      <TextCellText>{props.children}</TextCellText>
    </BaseCell>
  );
}

export default TextCell;

export function TextCellText(
  props: PropsWithChildren<Record<never, never>>
): React.ReactElement {
  return <Text className={styles.cellText}>{props.children}</Text>;
}
