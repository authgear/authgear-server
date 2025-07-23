import React, { useContext, useCallback, useEffect } from "react";
import FormTextField from "../../FormTextField";
import styles from "./ResourceForm.module.css";
import { Context } from "@oursky/react-messageformat";
import cn from "classnames";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import { useFormTopErrors } from "../../form";
import { useErrorMessageBarContext } from "../../ErrorMessageBar";
import { useLoading } from "../../hook/loading";
import PrimaryButton from "../../PrimaryButton";

export interface ResourceFormState {
  name: string;
  resourceURI: string;
}

export interface ResourceFormProps {
  className?: string;
  state: ResourceFormState;
  setState: (fn: (state: ResourceFormState) => ResourceFormState) => void;
}

export function sanitizeFormState(state: ResourceFormState): ResourceFormState {
  return {
    name: state.name.trim(),
    resourceURI: state.resourceURI.trim(),
  };
}

function isFormIncomplete(state: ResourceFormState): boolean {
  const s = sanitizeFormState(state);
  const incomplete = !s.name || !s.resourceURI;
  return !incomplete;
}

export const ResourceForm: React.VFC<ResourceFormProps> =
  function ResourceForm({ className, state, setState }) {
    const { renderToString } = useContext(Context);
    const handleNameChange = useCallback(
      (_e, value) => setState((s) => ({ ...s, name: value ?? "" })),
      [setState]
    );
    const handleResourceURIChange = useCallback(
      (_e, value) => setState((s) => ({ ...s, resourceURI: value ?? "" })),
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
            label={renderToString("ResourceForm.name.label")}
            description={renderToString("ResourceForm.name.description")}
            fieldName="name"
            parentJSONPointer=""
            type="text"
            value={state.name}
            onChange={handleNameChange}
          />
          <FormTextField
            required={true}
            label={renderToString("ResourceForm.resourceURI.label")}
            description={renderToString("ResourceForm.resourceURI.description")}
            fieldName="resourceURI"
            parentJSONPointer=""
            type="text"
            value={state.resourceURI}
            onChange={handleResourceURIChange}
            placeholder={renderToString("ResourceForm.resourceURI.placeholder")}
          />
        </div>
        <PrimaryButton
          type="submit"
          text={renderToString("create")}
          disabled={!canSave || !isFormIncomplete(state)}
        />
      </form>
    );
  };
