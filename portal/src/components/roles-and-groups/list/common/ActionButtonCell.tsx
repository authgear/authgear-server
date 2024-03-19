import React from "react";
import BaseCell from "./BaseCell";
import ActionButton from "../../../../ActionButton";
import styles from "./ActionButtonCell.module.css";
import { Text } from "@fluentui/react";
import { useSystemConfig } from "../../../../context/SystemConfigContext";
import cn from "classnames";

interface ActionButtonCellProps {
  text: string;
  onClick?: (e: any) => void;
  disabled?: boolean;
}

function ActionButtonCell(props: ActionButtonCellProps): React.ReactElement {
  const { text, onClick, disabled } = props;
  const { themes } = useSystemConfig();
  return (
    <BaseCell>
      <ActionButton
        text={
          <Text
            // TO BE CONFIRMED: How the themes color work for different states
            className={cn(
              styles.actionButtonText,
              disabled && "text-[#BFBFC3]"
            )}
            theme={themes.destructive}
          >
            {text}
          </Text>
        }
        className={styles.actionButton}
        theme={themes.destructive}
        onClick={onClick}
        disabled={disabled}
      />
    </BaseCell>
  );
}

export default ActionButtonCell;
