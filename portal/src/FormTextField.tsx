import React, { useContext, useMemo } from "react";
import { ITextFieldProps, TextField } from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";
import { useFormField } from "./form";
import { ErrorParseRule, renderErrors } from "./error/parse";

export interface FormTextFieldProps extends ITextFieldProps {
  parentJSONPointer: string | RegExp;
  fieldName: string;
  errorRules?: ErrorParseRule[];
}

const FormTextField: React.FC<FormTextFieldProps> = function FormTextField(
  props: FormTextFieldProps
) {
  const { parentJSONPointer, fieldName, errorRules, ...rest } = props;

  const { renderToString } = useContext(Context);

  const field = useMemo(
    () => ({
      parentJSONPointer,
      fieldName,
      rules: errorRules,
    }),
    [parentJSONPointer, fieldName, errorRules]
  );
  const { errors } = useFormField(field);
  const errorMessage = useMemo(
    () => renderErrors(errors, renderToString),
    [errors, renderToString]
  );

  return <TextField {...rest} errorMessage={errorMessage} />;
};

export default FormTextField;
