import React, { useContext, useCallback, useEffect } from "react";
import FormTextField from "../../FormTextField";
import styles from "./ScopeForm.module.css";
import { Context } from "../../intl";
import cn from "classnames";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import { useFormTopErrors } from "../../form";
import { useErrorMessageBarContext } from "../../ErrorMessageBar";
import { useLoading } from "../../hook/loading";
import PrimaryButton from "../../PrimaryButton";

export interface ScopeFormState {
  scope: string;
  description: string;
}

export interface ScopeFormProps {
  className?: string;
  mode: "create" | "edit";
  state: ScopeFormState;
  setState: (fn: (state: ScopeFormState) => ScopeFormState) => void;
}

export function sanitizeScopeFormState(state: ScopeFormState): ScopeFormState {
  return {
    scope: state.scope.trim(),
    description: state.description.trim(),
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
}) {
  const { renderToString } = useContext(Context);
  const handleDescriptionChange = useCallback(
    (_e, value) => setState((s) => ({ ...s, description: value ?? "" })),
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
