import { ITextFieldProps, TextField } from "@fluentui/react";
import React, { useCallback, useMemo } from "react";
import { useFormField } from "./error/FormFieldContext";
import FieldList from "./FieldList";
import cn from "classnames";
import styles from "./FormTextFieldList.module.scss";

interface TextFieldListItemProps {
  index: number;
  getJSONPointer: (index: number) => RegExp | string;
  parentJSONPointer: RegExp | string;
  textFieldProps?: ITextFieldProps;
  value: string;
  onChange: (value: string) => void;
}

const TextFieldListItem: React.FC<TextFieldListItemProps> = function TextFieldListItem(
  props: TextFieldListItemProps
) {
  const {
    index,
    parentJSONPointer,
    getJSONPointer,
    textFieldProps,
    value,
    onChange,
  } = props;

  const { value: _value, className: inputClassName, ...reducedTextFieldProps } =
    textFieldProps ?? {};

  const jsonPointer = useMemo(() => {
    return getJSONPointer(index);
  }, [index, getJSONPointer]);

  const { errorMessage } = useFormField(jsonPointer, parentJSONPointer, "");

  const _onChange = useCallback(
    (_event, newValue) => {
      if (newValue == null) {
        return;
      }
      onChange(newValue);
    },
    [onChange]
  );

  return (
    <TextField
      {...reducedTextFieldProps}
      className={cn(styles.inputField, inputClassName)}
      value={value}
      onChange={_onChange}
      errorMessage={errorMessage}
    />
  );
};

interface FormTextFieldListProps {
  className?: string;
  label?: React.ReactNode;
  jsonPointer: RegExp | string;
  parentJSONPointer: RegExp | string;
  fieldName: string;
  fieldNameMessageID?: string;
  getItemJSONPointer: (index: number) => RegExp | string;
  inputProps?: ITextFieldProps;
  list: string[];
  onListChange: (list: string[]) => void;
  addButtonLabelMessageID?: string;
  errorMessage?: string;
}

const FormTextFieldList: React.FC<FormTextFieldListProps> = function FormTextFieldList(
  props
) {
  const {
    className,
    label,
    jsonPointer,
    parentJSONPointer,
    fieldName,
    fieldNameMessageID,
    getItemJSONPointer,
    inputProps,
    list,
    onListChange,
    addButtonLabelMessageID,
    errorMessage,
  } = props;
  const makeDefaultItem = useCallback(() => "", []);
  const renderListItem = useCallback(
    (index: number, value: string, onChange: (newValue: string) => void) => (
      <TextFieldListItem
        index={index}
        getJSONPointer={getItemJSONPointer}
        parentJSONPointer={jsonPointer}
        textFieldProps={inputProps}
        value={value}
        onChange={onChange}
      />
    ),
    [getItemJSONPointer, inputProps, jsonPointer]
  );

  return (
    <FieldList
      className={className}
      label={label}
      jsonPointer={jsonPointer}
      parentJSONPointer={parentJSONPointer}
      fieldName={fieldName}
      fieldNameMessageID={fieldNameMessageID}
      list={list}
      onListChange={onListChange}
      makeDefaultItem={makeDefaultItem}
      renderListItem={renderListItem}
      addButtonLabelMessageID={addButtonLabelMessageID}
      errorMessage={errorMessage}
    />
  );
};

export default FormTextFieldList;
