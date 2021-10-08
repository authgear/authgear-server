import { ITextFieldProps, TextField } from "@fluentui/react";
import React, { useCallback, useContext, useMemo } from "react";
import FieldList from "./FieldList";
import cn from "classnames";
import { Context } from "@oursky/react-messageformat";
import styles from "./FormTextFieldList.module.scss";
import { useFormField } from "./form";
import { renderErrors } from "./error/parse";

interface TextFieldListItemProps {
  index: number;
  parentJSONPointer: string;
  textFieldProps?: ITextFieldProps;
  value: string;
  onChange: (value: string) => void;
}

const TextFieldListItem: React.FC<TextFieldListItemProps> =
  function TextFieldListItem(props: TextFieldListItemProps) {
    const { index, parentJSONPointer, textFieldProps, value, onChange } = props;
    const {
      value: _value,
      className: inputClassName,
      ...reducedTextFieldProps
    } = textFieldProps ?? {};

    const { renderToString } = useContext(Context);

    const field = useMemo(
      () => ({
        parentJSONPointer,
        fieldName: index.toString(10),
      }),
      [parentJSONPointer, index]
    );
    const { errors } = useFormField(field);
    const errorMessage = useMemo(
      () => renderErrors(errors, renderToString),
      [errors, renderToString]
    );

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
  parentJSONPointer: string;
  fieldName: string;
  inputProps?: ITextFieldProps;
  list: string[];
  onListChange: (list: string[]) => void;
  addButtonLabelMessageID?: string;
}

const FormTextFieldList: React.FC<FormTextFieldListProps> =
  function FormTextFieldList(props) {
    const {
      className,
      label,
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
          parentJSONPointer={`${parentJSONPointer}/${fieldName}`}
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
