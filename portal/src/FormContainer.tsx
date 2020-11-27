import React from "react";
import { useValidationError } from "./error/useValidationError";
import { FormContext } from "./error/FormContext";
import ShowUnhandledValidationErrorCause from "./error/ShowUnhandledValidationErrorCauses";
import ShowError from "./ShowError";

export interface FormContainerProps {
  error: unknown;
}

const FormContainer: React.FC<FormContainerProps> = function FormContainer(
  props
) {
  const {
    otherError,
    unhandledCauses,
    value: formContextValue,
  } = useValidationError(props.error);

  return (
    <FormContext.Provider value={formContextValue}>
      <ShowUnhandledValidationErrorCause causes={unhandledCauses} />
      {(unhandledCauses ?? []).length === 0 && otherError && (
        <ShowError error={otherError} />
      )}
      {props.children}
    </FormContext.Provider>
  );
};

export default FormContainer;
