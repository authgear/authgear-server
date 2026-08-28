import React, { useEffect, useContext, useCallback } from "react";
import cn from "classnames";
import { useLoading } from "../../hook/loading";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import { useErrorMessageBarContext } from "../../ErrorMessageBar";
import { useFormTopErrors } from "../../form";
import { Checkbox, Text } from "@fluentui/react";
import FormTextField from "../../FormTextField";
import PrimaryButton from "../../PrimaryButton";
import { Context as MessageContext, FormattedMessage } from "../../intl";

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
  // form says so -- same description pair ScopeForm shows on the edit path.
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
      (_e, value) => setState((s) => ({ ...s, scope: value ?? "" })),
      [setState]
    );
    const handleDescriptionChange = useCallback(
      (_e, value) => setState((s) => ({ ...s, description: value ?? "" })),
      [setState]
    );
    const handleAllowDynamicAccessChange = useCallback(
      (
        _e?: React.FormEvent<HTMLElement | HTMLInputElement>,
        checked?: boolean
      ) =>
        setState((s) => ({
          ...s,
          allowDynamicThirdPartyClientAccess: checked ?? false,
        })),
      [setState]
    );

    return (
      <form
        onSubmit={onSubmit}
        className={cn("flex flex-col max-w-200", className)}
      >
        {/* The scope/description fields keep their original single row with
            the Add button; h-22 reserves space for their validation errors. */}
        <div className="flex items-start gap-x-4 h-22">
          <FormTextField
            className="flex-1"
            required={true}
            label={renderToString("CreateScopeForm.scope.label")}
            fieldName="scope"
            parentJSONPointer=""
            type="text"
            value={state.scope}
            onChange={handleScopeChange}
            placeholder={renderToString("CreateScopeForm.scope.placeholder")}
          />
          <FormTextField
            className="flex-1"
            label={renderToString("CreateScopeForm.description.label")}
            fieldName="description"
            parentJSONPointer=""
            type="text"
            value={state.description}
            onChange={handleDescriptionChange}
            placeholder={renderToString(
              "CreateScopeForm.description.placeholder"
            )}
          />
          <PrimaryButton
            className="flex-none mt-[30px]"
            type="submit"
            text={renderToString("CreateScopeForm.add.button")}
            iconProps={{ iconName: "Add" }}
            disabled={!canSave || isFormIncomplete(state)}
          />
        </div>
        {/* Full width under the fields, mirroring ScopeForm's block on the
            edit path: same label, same two descriptions, same trigger. */}
        <div className="flex flex-col gap-1">
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
      </form>
    );
  };
