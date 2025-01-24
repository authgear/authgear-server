import React, { useMemo } from "react";
import { ITextFieldProps } from "@fluentui/react";
import { ErrorParseRule } from "./error/parse";
import { useErrorMessage } from "./formbinding";
import TextField from "./TextField";

export interface FormTextFieldProps extends ITextFieldProps {
  parentJSONPointer: string | RegExp;
  fieldName: string;
  errorRules?: ErrorParseRule[];
}

const FormTextField: React.VFC<FormTextFieldProps> = function FormTextField(
  props: FormTextFieldProps
) {
  const {
    parentJSONPointer,
    fieldName,
    errorRules,
    disabled: ownDisabled,
    ...rest
  } = props;
  const field = useMemo(
    () => ({
      parentJSONPointer,
      fieldName,
      rules: errorRules,
    }),
    [parentJSONPointer, fieldName, errorRules]
  );
  const { disabled: ctxDisabled, ...textFieldProps } = useErrorMessage(field);
  return (
    <TextField
      {...rest}
      {...textFieldProps}
      disabled={(ownDisabled ?? false) || ctxDisabled}
    />
  );
};

export default FormTextField;
