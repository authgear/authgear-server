import React, { useMemo } from "react";
import BaseCell from "./BaseCell";
import ActionButton from "../../../../ActionButton";
import styles from "./ActionButtonCell.module.css";
import { useSystemConfig } from "../../../../context/SystemConfigContext";

interface ActionButtonCellProps {
  text: string;
  onClick?: (e: any) => void;
  disabled?: boolean;
  variant?: "destructive" | "default";
}

function ActionButtonCell(props: ActionButtonCellProps): React.ReactElement {
  const { text, onClick, disabled, variant = "default" } = props;
  const { themes } = useSystemConfig();
  const theme = useMemo(() => {
    switch (variant) {
      case "destructive":
        return themes["destructive"];
      default:
        return themes["actionButton"];
    }
  }, [themes, variant]);

  return (
    <BaseCell>
      <ActionButton
        text={text}
        styles={{
          label: { fontWeight: 600 },
          labelDisabled: { color: "#BFBFC3" },
        }}
        className={styles.actionButton}
        theme={theme}
        onClick={onClick}
        disabled={disabled}
      />
    </BaseCell>
  );
}

export default ActionButtonCell;
