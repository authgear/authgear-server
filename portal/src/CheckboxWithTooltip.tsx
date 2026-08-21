import React, { useCallback } from "react";
import cn from "classnames";
import { Checkbox, Text } from "@radix-ui/themes";
import { InfoCircledIcon } from "@radix-ui/react-icons";

import styles from "./CheckboxWithTooltip.module.css";
import { Tooltip } from "./components/v2/Tooltip/Tooltip";
import { FormattedMessage } from "./intl";

interface CheckboxWithTooltipProps {
  className?: string;
  label?: React.ReactNode;
  checked?: boolean;
  disabled?: boolean;
  tooltipMessageId: string;
  tooltipMessageValues?: Record<string, any>;
  onCheckedChange?: (checked: boolean) => void;
}

const CheckboxWithTooltip: React.VFC<CheckboxWithTooltipProps> =
  function CheckboxWithTooltip(props: CheckboxWithTooltipProps) {
    const {
      className,
      label,
      checked,
      disabled,
      tooltipMessageId,
      tooltipMessageValues,
      onCheckedChange,
    } = props;

    const handleCheckedChange = useCallback(
      (checked: boolean | "indeterminate") => {
        if (checked === "indeterminate") {
          return;
        }
        onCheckedChange?.(checked);
      },
      [onCheckedChange]
    );

    // Keep the icon from toggling the checkbox via the surrounding label.
    const handleIconClick = useCallback((e: React.MouseEvent) => {
      e.preventDefault();
    }, []);

    return (
      <div className={cn(styles.root, className)}>
        <label className={styles.checkboxLabel}>
          <Checkbox
            checked={checked ?? false}
            disabled={disabled}
            onCheckedChange={handleCheckedChange}
          />
          <Text size="2">
            {label}
            <Tooltip
              content={
                <FormattedMessage
                  id={tooltipMessageId}
                  values={tooltipMessageValues}
                />
              }
            >
              <InfoCircledIcon
                className={styles.infoIcon}
                onClick={handleIconClick}
              />
            </Tooltip>
          </Text>
        </label>
      </div>
    );
  };

export default CheckboxWithTooltip;
