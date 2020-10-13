import React, { useContext, useMemo } from "react";
import { ITextFieldProps, TextField } from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";

import { useFormField } from "./error/FormFieldContext";

interface FormTextFieldProps extends ITextFieldProps {
  jsonPointer: RegExp | string;
  parentJSONPointer: RegExp | string;
  fieldName: string;
  fieldNameMessageID?: string;
}

const FormTextField: React.FC<FormTextFieldProps> = function FormTextField(
  props: FormTextFieldProps
) {
  const {
    jsonPointer,
    parentJSONPointer,
    fieldName,
    fieldNameMessageID,
    errorMessage: errorMessageProps,
    label: labelProps,
    ...rest
  } = props;

  const { renderToString } = useContext(Context);

  const { errorMessage } = useFormField(
    jsonPointer,
    parentJSONPointer,
    fieldName,
    fieldNameMessageID
  );

  const localizedFieldName = useMemo(() => {
    return fieldNameMessageID != null
      ? renderToString(fieldNameMessageID)
      : undefined;
  }, [renderToString, fieldNameMessageID]);

  return (
    <TextField
      {...rest}
      errorMessage={errorMessageProps ?? errorMessage}
      label={labelProps ?? localizedFieldName}
    />
  );
};

export default FormTextField;
