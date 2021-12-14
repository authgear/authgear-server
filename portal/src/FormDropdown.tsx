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
  const { parentJSONPointer, fieldName, errorRules, ...rest } = props;
  const field = useMemo(
    () => ({
      parentJSONPointer,
      fieldName,
      rules: errorRules,
    }),
    [parentJSONPointer, fieldName, errorRules]
  );
  const extraProps = useErrorMessageString(field);
  return <Dropdown {...rest} {...extraProps} />;
};

export default FormDropdown;
