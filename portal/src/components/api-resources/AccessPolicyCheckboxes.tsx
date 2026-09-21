import React, { useCallback, useContext, useId } from "react";
import cn from "classnames";
import { Checkbox, Text } from "@radix-ui/themes";
import { Context, FormattedMessage } from "../../intl";
import { FieldLabelWithTooltip } from "../v2/FieldLabelWithTooltip/FieldLabelWithTooltip";
import {
  VISIBLE_ACCESS_POLICY_KEYS,
  AccessPolicyState,
  VisibleAccessPolicyKey,
  accessPolicyDescriptionIDs,
  accessPolicyLabelIDs,
  withAccessPolicyKey,
} from "./accessPolicy";
import styles from "./AccessPolicyCheckboxes.module.css";

export interface AccessPolicyCheckboxesProps {
  className?: string;
  title: React.ReactNode;
  hint?: React.ReactNode;
  value: AccessPolicyState;
  disabled?: boolean;
  onChange: (value: AccessPolicyState) => void;
}

function AccessPolicyCheckbox({
  policyKey,
  checked,
  disabled,
  onChange,
}: {
  policyKey: VisibleAccessPolicyKey;
  checked: boolean;
  disabled?: boolean;
  onChange: (key: VisibleAccessPolicyKey, checked: boolean) => void;
}): React.ReactElement {
  const { renderToString } = useContext(Context);
  const id = useId();
  const onCheckedChange = useCallback(
    (next: boolean | "indeterminate") => {
      if (next === "indeterminate") {
        return;
      }
      onChange(policyKey, next);
    },
    [onChange, policyKey]
  );
  return (
    <div className={styles.item}>
      <Checkbox
        id={id}
        checked={checked}
        disabled={disabled}
        onCheckedChange={onCheckedChange}
      />
      <Text as="span" size="2">
        {/* The icon sits beside the label, not inside it, so hovering or
            clicking it does not toggle the checkbox. */}
        <FieldLabelWithTooltip
          tooltip={
            <FormattedMessage id={accessPolicyDescriptionIDs[policyKey]} />
          }
          tooltipLabel={renderToString(accessPolicyDescriptionIDs[policyKey])}
        >
          <label htmlFor={id}>
            <FormattedMessage id={accessPolicyLabelIDs[policyKey]} />
          </label>
        </FieldLabelWithTooltip>
      </Text>
    </div>
  );
}

// The client categories of an access policy as a checkbox group, used by the
// create-resource dialog. The scope forms use AccessPolicyChips, and the
// resource details page saves each category on toggle instead.
export function AccessPolicyCheckboxes({
  className,
  title,
  hint,
  value,
  disabled,
  onChange,
}: AccessPolicyCheckboxesProps): React.ReactElement {
  const titleID = useId();
  const onItemChange = useCallback(
    (key: VisibleAccessPolicyKey, checked: boolean) => {
      onChange(withAccessPolicyKey(value, key, checked));
    },
    [onChange, value]
  );
  return (
    <div
      role="group"
      aria-labelledby={titleID}
      className={cn(styles.root, className)}
    >
      <div className={styles.heading}>
        <Text
          as="p"
          id={titleID}
          size="2"
          weight="medium"
          className={styles.title}
        >
          {title}
        </Text>
        {hint != null ? (
          <Text as="p" size="1" color="gray" className={styles.hint}>
            {hint}
          </Text>
        ) : null}
      </div>
      <div className={styles.items}>
        {VISIBLE_ACCESS_POLICY_KEYS.map((key) => (
          <AccessPolicyCheckbox
            key={key}
            policyKey={key}
            checked={value[key]}
            disabled={disabled}
            onChange={onItemChange}
          />
        ))}
      </div>
    </div>
  );
}
