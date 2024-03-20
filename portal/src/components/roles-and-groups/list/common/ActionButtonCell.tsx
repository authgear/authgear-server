import React from "react";
import BaseCell from "./BaseCell";
import ActionButton from "../../../../ActionButton";
import styles from "./ActionButtonCell.module.css";
import { ITheme } from "@fluentui/react";
import { useSystemConfig } from "../../../../context/SystemConfigContext";

interface ActionButtonCellProps {
  text: string;
  onClick?: (e: any) => void;
  disabled?: boolean;
  theme?: ITheme;
}

function ActionButtonCell(props: ActionButtonCellProps): React.ReactElement {
  const { text, onClick, disabled, theme } = props;
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
        theme={theme ?? themes.destructive}
        onClick={onClick}
        disabled={disabled}
      />
    </BaseCell>
  );
}

export default ActionButtonCell;
