import React, { useEffect, useContext, useCallback } from "react";
import cn from "classnames";
import { Checkbox, Text } from "@radix-ui/themes";
import { PlusIcon } from "@radix-ui/react-icons";
import { useLoading } from "../../hook/loading";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import { useErrorMessageBarContext } from "../../ErrorMessageBar";
import { useFormTopErrors } from "../../form";
import { Context as MessageContext, FormattedMessage } from "../../intl";
import { PrimaryButton } from "../v2/Button/PrimaryButton/PrimaryButton";
import { TextField } from "../v2/TextField/TextField";
import styles from "./CreateScopeForm.module.css";

export interface CreateScopeFormState {
  scope: string;
  description: string;
  allowDynamicThirdPartyClientAccess: boolean;
}

export interface CreateScopeFormProps {
  className?: string;
  state: CreateScopeFormState;
  setState: (fn: (state: CreateScopeFormState) => CreateScopeFormState) => void;
  // Whether the parent resource allows dynamic third-party client access.
  // When it does not, the scope-level checkbox has no effect yet, and the
  // form says so -- same description pair EditScopeDialog shows on the edit
  // path.
  resourceAllowsDynamicAccess: boolean;
}

export function sanitizeCreateScopeFormState(
  state: CreateScopeFormState
): CreateScopeFormState {
  return {
    scope: state.scope.trim(),
    description: state.description.trim(),
    allowDynamicThirdPartyClientAccess:
      state.allowDynamicThirdPartyClientAccess,
  };
}

function isFormIncomplete(state: CreateScopeFormState): boolean {
  const s = sanitizeCreateScopeFormState(state);
  return !s.scope;
}

export const CreateScopeForm: React.VFC<CreateScopeFormProps> =
  function CreateScopeForm({
    className,
    state,
    setState,
    resourceAllowsDynamicAccess,
  }) {
    const { renderToString } = useContext(MessageContext);
    const { onSubmit, canSave, isUpdating } = useFormContainerBaseContext();
    useLoading(isUpdating);
    const errors = useFormTopErrors();
    const { setErrors } = useErrorMessageBarContext();
    useEffect(() => {
      setErrors(errors);
    }, [errors, setErrors]);

    const handleScopeChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const scope = e.target.value;
        setState((s) => ({ ...s, scope }));
      },
      [setState]
    );
    const handleDescriptionChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const description = e.target.value;
        setState((s) => ({ ...s, description }));
      },
      [setState]
    );
    const handleAllowDynamicAccessChange = useCallback(
      (checked: boolean | "indeterminate") => {
        if (checked === "indeterminate") {
          return;
        }
        setState((s) => ({
          ...s,
          allowDynamicThirdPartyClientAccess: checked,
        }));
      },
      [setState]
    );

    return (
      <form onSubmit={onSubmit} className={cn(styles.form, className)}>
        <div className={styles.root}>
          <div className={styles.field}>
            <TextField
              size="2"
              required={true}
              label={<FormattedMessage id="CreateScopeForm.scope.label" />}
              fieldName="scope"
              parentJSONPointer=""
              type="text"
              value={state.scope}
              onChange={handleScopeChange}
              placeholder={renderToString("CreateScopeForm.scope.placeholder")}
            />
          </div>
          <div className={styles.field}>
            <TextField
              size="2"
              label={
                <FormattedMessage id="CreateScopeForm.description.label" />
              }
              fieldName="description"
              parentJSONPointer=""
              type="text"
              value={state.description}
              onChange={handleDescriptionChange}
              placeholder={renderToString(
                "CreateScopeForm.description.placeholder"
              )}
            />
          </div>
          <div className={styles.submit}>
            <PrimaryButton
              size="2"
              type="submit"
              text={
                <span className={styles.submitContent}>
                  <PlusIcon width="1rem" height="1rem" />
                  <FormattedMessage id="CreateScopeForm.add.button" />
                </span>
              }
              disabled={!canSave || isFormIncomplete(state)}
              loading={isUpdating}
            />
          </div>
        </div>
        {/* Full width under the fields, mirroring EditScopeDialog's block on
            the edit path: same label, same two descriptions, same trigger. */}
        <div className={styles.dynamicAccess}>
          <label className={styles.dynamicAccessLabel}>
            <Checkbox
              checked={state.allowDynamicThirdPartyClientAccess}
              onCheckedChange={handleAllowDynamicAccessChange}
            />
            <Text size="2">
              <FormattedMessage id="ScopeForm.allow-dynamic-access.label" />
            </Text>
          </label>
          <Text as="p" size="1" color="gray">
            {resourceAllowsDynamicAccess ? (
              <FormattedMessage id="ScopeForm.allow-dynamic-access.description" />
            ) : (
              <FormattedMessage id="ScopeForm.allow-dynamic-access.resource-off" />
            )}
          </Text>
        </div>
      </form>
    );
  };
