import React, { PropsWithChildren } from "react";
import BaseCell from "./BaseCell";
import styles from "./DescriptionCell.module.css";
import { Text } from "@fluentui/react";

function DescriptionCell(
  props: PropsWithChildren<Record<never, never>>
): React.ReactElement {
  return (
    <BaseCell>
      <Text className={styles.description}>{props.children}</Text>
    </BaseCell>
  );
}

export default DescriptionCell;
