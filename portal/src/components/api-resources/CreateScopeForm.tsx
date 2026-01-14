import React, { useEffect, useContext, useCallback } from "react";
import cn from "classnames";
import { useLoading } from "../../hook/loading";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import { useErrorMessageBarContext } from "../../ErrorMessageBar";
import { useFormTopErrors } from "../../form";
import FormTextField from "../../FormTextField";
import PrimaryButton from "../../PrimaryButton";
import { Context as MessageContext } from "../../intl";

export interface CreateScopeFormState {
  scope: string;
  description: string;
}

export interface CreateScopeFormProps {
  className?: string;
  state: CreateScopeFormState;
  setState: (fn: (state: CreateScopeFormState) => CreateScopeFormState) => void;
}

export function sanitizeCreateScopeFormState(
  state: CreateScopeFormState
): CreateScopeFormState {
  return {
    scope: state.scope.trim(),
    description: state.description.trim(),
  };
}

function isFormIncomplete(state: CreateScopeFormState): boolean {
  const s = sanitizeCreateScopeFormState(state);
  return !s.scope;
}

export const CreateScopeForm: React.VFC<CreateScopeFormProps> =
  function CreateScopeForm({ className, state, setState }) {
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

    return (
      <form
        onSubmit={onSubmit}
        className={cn("flex items-start max-w-200 gap-x-4 h-22", className)}
      >
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
      </form>
    );
  };
