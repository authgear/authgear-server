import React, { useContext, useMemo } from "react";
import { ITextFieldProps, TextField } from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";
import { useFormField } from "./form";
import { ErrorParseRule, renderErrors } from "./error/parse";

export interface FormTextFieldProps extends ITextFieldProps {
  parentJSONPointer: string;
  fieldName: string;
  fieldNameMessageID?: string;
  errorRules?: ErrorParseRule[];
  hideLabel?: boolean;
}

const FormTextField: React.FC<FormTextFieldProps> = function FormTextField(
  props: FormTextFieldProps
) {
  const {
    parentJSONPointer,
    fieldName,
    fieldNameMessageID,
    errorRules,
    label: labelProps,
    hideLabel,
    ...rest
  } = props;

  const { renderToString } = useContext(Context);

  const field = useMemo(
    () => ({
      parentJSONPointer,
      fieldName,
      fieldNameMessageID,
      rules: errorRules,
    }),
    [parentJSONPointer, fieldName, fieldNameMessageID, errorRules]
  );
  const { errors } = useFormField(field);
  const errorMessage = useMemo(
    () => renderErrors(field, errors, renderToString),
    [field, errors, renderToString]
  );

  const localizedFieldName = useMemo(() => {
    return fieldNameMessageID != null
      ? renderToString(fieldNameMessageID)
      : undefined;
  }, [renderToString, fieldNameMessageID]);

  return (
    <TextField
      {...rest}
      errorMessage={errorMessage}
      label={hideLabel ? undefined : labelProps ?? localizedFieldName}
    />
  );
};

export default FormTextField;
