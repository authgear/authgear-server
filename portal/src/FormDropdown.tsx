import React, { useMemo } from "react";
import { IDropdownProps, Dropdown } from "@fluentui/react";
import { ErrorParseRule } from "./error/parse";
import { useErrorMessageString } from "./formbinding";

export interface FormDropdownProps extends IDropdownProps {
  parentJSONPointer: string | RegExp;
  fieldName: string;
  errorRules?: ErrorParseRule[];
}

const FormDropdown: React.FC<FormDropdownProps> = function FormDropdown(
  props: FormDropdownProps
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
  const { disabled: ctxDisabled, ...extraProps } = useErrorMessageString(field);
  return (
    <Dropdown {...rest} {...extraProps} disabled={ownDisabled ?? ctxDisabled} />
  );
};

export default FormDropdown;
