import React, { useEffect, useContext, useCallback, useMemo } from "react";
import cn from "classnames";
import { PlusIcon } from "@radix-ui/react-icons";
import { useLoading } from "../../hook/loading";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import { useErrorMessageBarContext } from "../../ErrorMessageBar";
import { useFormTopErrors } from "../../form";
import { Context as MessageContext, FormattedMessage } from "../../intl";
import { PrimaryButton } from "../v2/Button/PrimaryButton/PrimaryButton";
import { TextField } from "../v2/TextField/TextField";
import { FieldLabelWithTooltip } from "../v2/FieldLabelWithTooltip/FieldLabelWithTooltip";
import { FormField } from "../v2/FormField/FormField";
import { AccessPolicyChips } from "./AccessPolicyChips";
import { Callout } from "../v2/Callout/Callout";
import { AccessPolicy } from "../../graphql/adminapi/globalTypes.generated";
import {
  AccessPolicyState,
  unreachableAccessPolicyKeys,
  useAccessPolicyCategoryList,
} from "./accessPolicy";
import styles from "./CreateScopeForm.module.css";

export interface CreateScopeFormState {
  scope: string;
  description: string;
  accessPolicy: AccessPolicyState;
}

export interface CreateScopeFormProps {
  className?: string;
  // The parent resource's policy: a category allowed here but not there is
  // unreachable, and is warned about.
  resourceAccessPolicy: AccessPolicy;
  state: CreateScopeFormState;
  setState: (fn: (state: CreateScopeFormState) => CreateScopeFormState) => void;
}

export function sanitizeCreateScopeFormState(
  state: CreateScopeFormState
): CreateScopeFormState {
  return {
    scope: state.scope.trim(),
    description: state.description.trim(),
    accessPolicy: state.accessPolicy,
  };
}

function isFormIncomplete(state: CreateScopeFormState): boolean {
  const s = sanitizeCreateScopeFormState(state);
  return !s.scope;
}

export const CreateScopeForm: React.VFC<CreateScopeFormProps> =
  function CreateScopeForm({
    className,
    resourceAccessPolicy,
    state,
    setState,
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
    const handleAccessPolicyChange = useCallback(
      (accessPolicy: AccessPolicyState) => {
        setState((s) => ({ ...s, accessPolicy }));
      },
      [setState]
    );

    const formatCategories = useAccessPolicyCategoryList();
    const unreachableKeys = useMemo(
      () =>
        unreachableAccessPolicyKeys(state.accessPolicy, resourceAccessPolicy),
      [state.accessPolicy, resourceAccessPolicy]
    );

    const descriptionLabel = (
      <FieldLabelWithTooltip
        tooltip={<FormattedMessage id="ScopeForm.description.tooltip" />}
        tooltipLabel={renderToString("ScopeForm.description.tooltip")}
      >
        <FormattedMessage id="CreateScopeForm.description.label" />
      </FieldLabelWithTooltip>
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
              label={descriptionLabel}
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
        </div>
        <FormField
          size="2"
          labelSpace="1"
          label={
            <FieldLabelWithTooltip
              tooltip={
                <FormattedMessage id="AccessPolicyCheckboxes.scope.hint" />
              }
              tooltipLabel={renderToString("AccessPolicyCheckboxes.scope.hint")}
            >
              <FormattedMessage id="AccessPolicyCheckboxes.scope.title" />
            </FieldLabelWithTooltip>
          }
        >
          <AccessPolicyChips
            value={state.accessPolicy}
            onChange={handleAccessPolicyChange}
          />
        </FormField>
        {unreachableKeys.length > 0 ? (
          <Callout
            type="warning"
            size="1"
            showCloseButton={false}
            text={
              <FormattedMessage
                id="ScopeForm.access-policy.unreachable"
                values={{ categories: formatCategories(unreachableKeys) }}
              />
            }
          />
        ) : null}
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
      </form>
    );
  };
