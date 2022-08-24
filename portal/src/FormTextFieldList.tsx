import React, { useCallback, useMemo } from "react";
import FieldList from "./FieldList";
import cn from "classnames";
import styles from "./FormTextFieldList.module.css";
import { useFormField } from "./form";
import { joinParentChild } from "./util/jsonpointer";
import ErrorRenderer from "./ErrorRenderer";
import TextField, { TextFieldProps } from "./TextField";

interface TextFieldListItemProps {
  index: number;
  parentJSONPointer: string | RegExp;
  textFieldProps?: TextFieldProps;
  value: string;
  onChange: (value: string) => void;
}

const TextFieldListItem: React.VFC<TextFieldListItemProps> =
  function TextFieldListItem(props: TextFieldListItemProps) {
    const { index, parentJSONPointer, textFieldProps, value, onChange } = props;
    const {
      value: _value,
      className: inputClassName,
      ...reducedTextFieldProps
    } = textFieldProps ?? {};

    const field = useMemo(
      () => ({
        parentJSONPointer,
        fieldName: index.toString(10),
      }),
      [parentJSONPointer, index]
    );
    const { errors } = useFormField(field);

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
        errorMessage={
          errors.length > 0 ? <ErrorRenderer errors={errors} /> : undefined
        }
      />
    );
  };

export interface FormTextFieldListProps {
  className?: string;
  label?: React.ReactNode;
  description?: string;
  parentJSONPointer: string | RegExp;
  fieldName: string;
  inputProps?: TextFieldProps;
  list: string[];
  onListChange: (list: string[]) => void;
  addButtonLabelMessageID?: string;
}

const FormTextFieldList: React.VFC<FormTextFieldListProps> =
  function FormTextFieldList(props) {
    const {
      className,
      label,
      description,
      parentJSONPointer,
      fieldName,
      inputProps,
      list,
      onListChange,
      addButtonLabelMessageID,
    } = props;
    const makeDefaultItem = useCallback(() => "", []);
    const renderListItem = useCallback(
      (index: number, value: string, onChange: (newValue: string) => void) => (
        <TextFieldListItem
          index={index}
          parentJSONPointer={joinParentChild(parentJSONPointer, fieldName)}
          textFieldProps={inputProps}
          value={value}
          onChange={onChange}
        />
      ),
      [inputProps, parentJSONPointer, fieldName]
    );

    return (
      <FieldList
        className={className}
        label={label}
        description={description}
        parentJSONPointer={parentJSONPointer}
        fieldName={fieldName}
        list={list}
        onListChange={onListChange}
        makeDefaultItem={makeDefaultItem}
        renderListItem={renderListItem}
        addButtonLabelMessageID={addButtonLabelMessageID}
      />
    );
  };

export default FormTextFieldList;
