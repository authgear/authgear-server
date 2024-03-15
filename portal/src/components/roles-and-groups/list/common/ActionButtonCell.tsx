import React from "react";
import BaseCell from "./BaseCell";
import ActionButton from "../../../../ActionButton";
import styles from "./ActionButtonCell.module.css";
import { Text } from "@fluentui/react";
import { useSystemConfig } from "../../../../context/SystemConfigContext";

interface ActionButtonCellProps {
  text: string;
  onClick?: (e: any) => void;
}

function ActionButtonCell(props: ActionButtonCellProps): React.ReactElement {
  const { text, onClick } = props;
  const { themes } = useSystemConfig();
  return (
    <BaseCell>
      <ActionButton
        text={
          <Text className={styles.actionButtonText} theme={themes.destructive}>
            {text}
          </Text>
        }
        className={styles.actionButton}
        theme={themes.destructive}
        onClick={onClick}
      />
    </BaseCell>
  );
}

export default ActionButtonCell;
