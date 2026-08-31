import React, { useContext, useCallback, useEffect } from "react";
import { Checkbox, Text } from "@fluentui/react";
import FormTextField from "../../FormTextField";
import styles from "./ScopeForm.module.css";
import { Context, FormattedMessage } from "../../intl";
import cn from "classnames";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import { useFormTopErrors } from "../../form";
import { useErrorMessageBarContext } from "../../ErrorMessageBar";
import { useLoading } from "../../hook/loading";
import PrimaryButton from "../../PrimaryButton";

export interface ScopeFormState {
  scope: string;
  description: string;
  allowDynamicThirdPartyClientAccess: boolean;
}

export interface ScopeFormProps {
  className?: string;
  mode: "create" | "edit";
  state: ScopeFormState;
  setState: (fn: (state: ScopeFormState) => ScopeFormState) => void;
  // Whether the parent resource allows dynamic third-party client access.
  // When it does not, the scope-level checkbox has no effect yet, and the
  // form shows a hint saying so.
  resourceAllowsDynamicAccess: boolean;
}

export function sanitizeScopeFormState(state: ScopeFormState): ScopeFormState {
  return {
    scope: state.scope.trim(),
    description: state.description.trim(),
    allowDynamicThirdPartyClientAccess:
      state.allowDynamicThirdPartyClientAccess,
  };
}

function isFormIncomplete(state: ScopeFormState): boolean {
  const s = sanitizeScopeFormState(state);
  return !s.scope;
}

export const ScopeForm: React.VFC<ScopeFormProps> = function ScopeForm({
  className,
  state,
  setState,
  mode,
  resourceAllowsDynamicAccess,
}) {
  const { renderToString } = useContext(Context);
  const handleDescriptionChange = useCallback(
    (_e, value) => setState((s) => ({ ...s, description: value ?? "" })),
    [setState]
  );
  const handleAllowDynamicAccessChange = useCallback(
    (_e?: React.FormEvent<HTMLElement | HTMLInputElement>, checked?: boolean) =>
      setState((s) => ({
        ...s,
        allowDynamicThirdPartyClientAccess: checked ?? false,
      })),
    [setState]
  );
  const { onSubmit, canSave, isUpdating } = useFormContainerBaseContext();

  useLoading(isUpdating);

  const errors = useFormTopErrors();
  const { setErrors } = useErrorMessageBarContext();
  useEffect(() => {
    setErrors(errors);
  }, [errors, setErrors]);

  return (
    <form className={cn(styles.root, className)} onSubmit={onSubmit}>
      <div className={styles.formFields}>
        <FormTextField
          required={true}
          label={renderToString("ScopeForm.scope.label")}
          fieldName="scope"
          parentJSONPointer=""
          type="text"
          value={state.scope}
          readOnly={true}
        />
        <FormTextField
          label={renderToString("ScopeForm.description.label")}
          fieldName="description"
          parentJSONPointer=""
          type="text"
          value={state.description}
          onChange={handleDescriptionChange}
        />
        <div className={styles.dynamicAccess}>
          <Checkbox
            label={renderToString("ScopeForm.allow-dynamic-access.label")}
            checked={state.allowDynamicThirdPartyClientAccess}
            onChange={handleAllowDynamicAccessChange}
          />
          <Text
            variant="small"
            block={true}
            styles={{ root: { color: "var(--gray-11)" } }}
          >
            {resourceAllowsDynamicAccess ? (
              <FormattedMessage id="ScopeForm.allow-dynamic-access.description" />
            ) : (
              <FormattedMessage id="ScopeForm.allow-dynamic-access.resource-off" />
            )}
          </Text>
        </div>
      </div>
      <PrimaryButton
        type="submit"
        text={
          mode === "edit" ? renderToString("save") : renderToString("create")
        }
        disabled={!canSave || isFormIncomplete(state)}
      />
    </form>
  );
};
