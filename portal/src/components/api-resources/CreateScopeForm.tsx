import React, { useEffect, useContext, useCallback } from "react";
import cn from "classnames";
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

    return (
      <form onSubmit={onSubmit} className={cn(styles.root, className)}>
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
            label={<FormattedMessage id="CreateScopeForm.description.label" />}
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
      </form>
    );
  };
