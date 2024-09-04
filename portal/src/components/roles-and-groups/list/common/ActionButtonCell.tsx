import React, { useMemo } from "react";
import cn from "classnames";
import BaseCell from "./BaseCell";
import ActionButton from "../../../../ActionButton";
import styles from "./ActionButtonCell.module.css";
import { useSystemConfig } from "../../../../context/SystemConfigContext";
import { Text } from "@fluentui/react";

interface ActionButtonCellProps {
  text: string;
  onClick?: (e: any) => void;
  disabled?: boolean;
  variant?: "destructive" | "default" | "no-action";
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

  switch (variant) {
    case "no-action":
      return (
        <BaseCell>
          {/* mx-1 to align with action button default padding 
          ref https://github.com/microsoft/fluentui/blob/4831884340f715d5a8d285e6862e19e85032b738/packages/react/src/components/Button/ActionButton/ActionButton.styles.ts#L14
         */}
          <Text className={cn(styles.actionButton, "mx-1")}>{text}</Text>
        </BaseCell>
      );
    default:
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
}

export default ActionButtonCell;
