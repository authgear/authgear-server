import React, { useMemo } from "react";
import { ITextFieldProps, TextField } from "@fluentui/react";
import { ErrorParseRule } from "./error/parse";
import { useErrorMessage } from "./formbinding";

export interface FormTextFieldProps extends ITextFieldProps {
  parentJSONPointer: string | RegExp;
  fieldName: string;
  errorRules?: ErrorParseRule[];
}

const FormTextField: React.FC<FormTextFieldProps> = function FormTextField(
  props: FormTextFieldProps
) {
  const { parentJSONPointer, fieldName, errorRules, ...rest } = props;
  const field = useMemo(
    () => ({
      parentJSONPointer,
      fieldName,
      rules: errorRules,
    }),
    [parentJSONPointer, fieldName, errorRules]
  );
  const textFieldProps = useErrorMessage(field);
  return <TextField {...rest} {...textFieldProps} />;
};

export default FormTextField;
