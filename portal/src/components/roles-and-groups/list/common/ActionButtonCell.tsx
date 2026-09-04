import React from "react";
import cn from "classnames";
import BaseCell from "./BaseCell";
import styles from "./ActionButtonCell.module.css";
import { Button, IconButton, Text } from "@radix-ui/themes";

interface ActionButtonCellProps {
  text?: string;
  icon?: React.ReactNode;
  ariaLabel?: string;
  onClick?: (e: any) => void;
  disabled?: boolean;
  variant?: "destructive" | "default" | "no-action";
}

function ActionButtonCell(props: ActionButtonCellProps): React.ReactElement {
  const {
    text,
    icon,
    ariaLabel,
    onClick,
    disabled,
    variant = "default",
  } = props;

  switch (variant) {
    case "no-action":
      return (
        <BaseCell>
          {/* mx-1 to align with action button default padding 
          ref https://github.com/microsoft/fluentui/blob/4831884340f715d5a8d285e6862e19e85032b738/packages/react/src/components/Button/ActionButton/ActionButton.styles.ts#L14
         */}
          <Text size="2" className={cn(styles.actionButton, "mx-1")}>
            {text}
          </Text>
        </BaseCell>
      );
    default:
      if (icon != null) {
        return (
          <BaseCell>
            <div className={styles.iconButtonContainer}>
              <IconButton
                size="2"
                variant="ghost"
                color={variant === "destructive" ? "red" : "gray"}
                className={styles.iconButton}
                aria-label={ariaLabel}
                onClick={onClick}
                disabled={disabled}
              >
                {icon}
              </IconButton>
            </div>
          </BaseCell>
        );
      }
      return (
        <BaseCell>
          <Button
            size="1"
            variant="ghost"
            color={variant === "destructive" ? "red" : "gray"}
            className={styles.actionButton}
            onClick={onClick}
            disabled={disabled}
          >
            {text}
          </Button>
        </BaseCell>
      );
  }
}

export default ActionButtonCell;
