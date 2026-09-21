import React, { useCallback } from "react";
import cn from "classnames";
import { CheckIcon, MinusIcon } from "@radix-ui/react-icons";
import { FormattedMessage } from "../../intl";
import {
  VISIBLE_ACCESS_POLICY_KEYS,
  AccessPolicyState,
  VisibleAccessPolicyKey,
  accessPolicyLabelIDs,
  withAccessPolicyKey,
} from "./accessPolicy";
import styles from "./AccessPolicyChips.module.css";

export interface AccessPolicyChipsProps {
  className?: string;
  value: AccessPolicyState;
  // Categories whose state differs across the items being edited. A mixed
  // chip reads as neither on nor off; clicking it turns it on.
  mixedKeys?: ReadonlySet<VisibleAccessPolicyKey>;
  disabled?: boolean;
  onChange: (value: AccessPolicyState) => void;
}

function Chip({
  policyKey,
  checked,
  mixed,
  disabled,
  onToggle,
}: {
  policyKey: VisibleAccessPolicyKey;
  checked: boolean;
  mixed: boolean;
  disabled?: boolean;
  onToggle: (key: VisibleAccessPolicyKey, checked: boolean) => void;
}): React.ReactElement {
  const onClick = useCallback(() => {
    onToggle(policyKey, mixed ? true : !checked);
  }, [onToggle, policyKey, checked, mixed]);
  return (
    <button
      type="button"
      role="checkbox"
      aria-checked={mixed ? "mixed" : checked}
      disabled={disabled}
      className={cn(
        styles.chip,
        !mixed && checked && styles["chip--checked"],
        mixed && styles["chip--mixed"]
      )}
      onClick={onClick}
    >
      {mixed ? (
        <MinusIcon className={styles.chipIcon} />
      ) : checked ? (
        <CheckIcon className={styles.chipIcon} />
      ) : null}
      <FormattedMessage id={accessPolicyLabelIDs[policyKey]} />
    </button>
  );
}

// The visible client categories as pill toggles that sit at text-field
// height, for forms where the access choice shares a row with other fields.
export function AccessPolicyChips({
  className,
  value,
  mixedKeys,
  disabled,
  onChange,
}: AccessPolicyChipsProps): React.ReactElement {
  const onToggle = useCallback(
    (key: VisibleAccessPolicyKey, checked: boolean) => {
      onChange(withAccessPolicyKey(value, key, checked));
    },
    [onChange, value]
  );
  return (
    <div role="group" className={cn(styles.root, className)}>
      {VISIBLE_ACCESS_POLICY_KEYS.map((key) => (
        <Chip
          key={key}
          policyKey={key}
          checked={value[key]}
          mixed={mixedKeys?.has(key) ?? false}
          disabled={disabled}
          onToggle={onToggle}
        />
      ))}
    </div>
  );
}
