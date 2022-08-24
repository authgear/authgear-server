import React, { useMemo } from "react";
import { ErrorParseRule } from "./error/parse";
import { useErrorMessage } from "./formbinding";
import PhoneTextField, { PhoneTextFieldProps } from "./PhoneTextField";

export interface FormPhoneTextFieldProps extends PhoneTextFieldProps {
  parentJSONPointer: string | RegExp;
  fieldName: string;
  errorRules?: ErrorParseRule[];
}

const FormPhoneTextField: React.VFC<FormPhoneTextFieldProps> =
  function FormPhoneTextField(props: FormPhoneTextFieldProps) {
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
      <PhoneTextField
        {...rest}
        {...textFieldProps}
        disabled={ownDisabled ?? ctxDisabled}
      />
    );
  };

export default FormPhoneTextField;
