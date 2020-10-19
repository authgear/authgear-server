import React from "react";
import {
  IStackProps,
  ITagPickerProps,
  Stack,
  TagPicker,
  Text,
} from "@fluentui/react";
import { useFormField } from "./error/FormFieldContext";

import styles from "./FormTagPicker.module.scss";

interface FormTagPickerProps extends ITagPickerProps {
  className?: string;
  inputClassName?: string;
  stackProps?: IStackProps;
  jsonPointer: RegExp | string;
  parentJSONPointer: RegExp | string;
  fieldName: string;
  fieldNameMessageID?: string;
  errorMessage?: string;
  label?: React.ReactNode;
  hideLabel?: boolean;
}

const FormTagPicker: React.FC<FormTagPickerProps> = function FormTagPicker(
  props: FormTagPickerProps
) {
  const {
    className,
    stackProps,
    inputClassName,
    jsonPointer,
    parentJSONPointer,
    fieldName,
    fieldNameMessageID,
    errorMessage: errorMessageProps,
    label,
    hideLabel,
    ...rest
  } = props;

  const { errorMessage } = useFormField(
    jsonPointer,
    parentJSONPointer,
    fieldName,
    fieldNameMessageID
  );

  return (
    <Stack {...stackProps} className={className}>
      {label}
      <TagPicker {...rest} className={inputClassName} />
      <Text className={styles.errorMessage}>
        {errorMessageProps ?? errorMessage}
      </Text>
    </Stack>
  );
};

export default FormTagPicker;
