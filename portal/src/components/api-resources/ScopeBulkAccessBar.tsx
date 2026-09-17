import React from "react";
import cn from "classnames";
import { Button, Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import { SecondaryButton } from "../v2/Button/SecondaryButton/SecondaryButton";
import styles from "./ScopeBulkAccessBar.module.css";

export interface ScopeBulkAccessBarProps {
  className?: string;
  count: number;
  disabled?: boolean;
  onChangeAccess: () => void;
  onClear: () => void;
}

// Appears beside the search field once at least one scope is selected.
export function ScopeBulkAccessBar({
  className,
  count,
  disabled,
  onChangeAccess,
  onClear,
}: ScopeBulkAccessBarProps): React.ReactElement | null {
  if (count === 0) {
    return null;
  }
  return (
    <div className={cn(styles.bar, className)}>
      <Text size="2" weight="medium">
        <FormattedMessage id="ScopeBulkAccessBar.selected" values={{ count }} />
      </Text>
      <SecondaryButton
        size="2"
        text={<FormattedMessage id="ScopeBulkAccessBar.change" />}
        disabled={disabled}
        onClick={onChangeAccess}
      />
      <Button
        type="button"
        variant="ghost"
        color="gray"
        size="2"
        disabled={disabled}
        onClick={onClear}
      >
        <FormattedMessage id="ScopeBulkAccessBar.clear" />
      </Button>
    </div>
  );
}
