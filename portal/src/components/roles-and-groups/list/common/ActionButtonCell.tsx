import React from "react";
import BaseCell from "./BaseCell";
import ActionButton from "../../../../ActionButton";
import styles from "./ActionButtonCell.module.css";
import { useSystemConfig } from "../../../../context/SystemConfigContext";

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
        text={text}
        styles={{
          label: { fontWeight: 600 },
          labelDisabled: { color: "#BFBFC3" },
        }}
        className={styles.actionButton}
        theme={themes.destructive}
        onClick={onClick}
        disabled={disabled}
      />
    </BaseCell>
  );
}

export default ActionButtonCell;
