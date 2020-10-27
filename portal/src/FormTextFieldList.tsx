import React, { useCallback, useMemo } from "react";
import {
  ActionButton,
  IconButton,
  ITextFieldProps,
  Stack,
  Text,
  TextField,
} from "@fluentui/react";
import cn from "classnames";
import produce from "immer";
import { FormattedMessage } from "@oursky/react-messageformat";

import { useSystemConfig } from "./context/SystemConfigContext";
import { useFormField } from "./error/FormFieldContext";

import styles from "./FormTextFieldList.module.scss";

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

interface TextFieldListItemProps {
  index: number;
  getJSONPointer: (index: number) => RegExp | string;
  parentJSONPointer: RegExp | string;
  textFieldProps?: ITextFieldProps;
  value: string;
  onChange: (index: number, value: string) => void;
  deleteItem: (index: number) => void;
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
    deleteItem,
  } = props;

  const { value: _value, className: inputClassName, ...reducedTextFieldProps } =
    textFieldProps ?? {};

  const jsonPointer = useMemo(() => {
    return getJSONPointer(index);
  }, [index, getJSONPointer]);

  const { themes } = useSystemConfig();
  const { errorMessage } = useFormField(jsonPointer, parentJSONPointer, "");

  const _onChange = useCallback(
    (_event, newValue) => {
      if (newValue == null) {
        return;
      }
      onChange(index, newValue);
    },
    [onChange, index]
  );

  const _onDeleteClick = useCallback(() => {
    deleteItem(index);
  }, [index, deleteItem]);

  return (
    <div className={styles.listItem}>
      <TextField
        {...reducedTextFieldProps}
        className={cn(styles.inputField, inputClassName)}
        value={value}
        onChange={_onChange}
        errorMessage={errorMessage}
      />
      <IconButton
        className={styles.deleteButton}
        onClick={_onDeleteClick}
        iconProps={{ iconName: "Delete" }}
        theme={themes.destructive}
      />
    </div>
  );
};

const FormTextFieldList: React.FC<FormTextFieldListProps> = function FormTextFieldList(
  props: FormTextFieldListProps
) {
  const {
    className,
    label,
    jsonPointer,
    getItemJSONPointer,
    parentJSONPointer,
    fieldName,
    fieldNameMessageID,
    inputProps,
    list,
    onListChange,
    addButtonLabelMessageID,
    errorMessage: errorMessageProps,
  } = props;

  const { themes } = useSystemConfig();

  const { errorMessage } = useFormField(
    jsonPointer,
    parentJSONPointer,
    fieldName,
    fieldNameMessageID
  );

  const onTextFieldChange = useCallback(
    (index: number, newValue: string) => {
      onListChange(
        produce(list, (draftList) => {
          draftList[index] = newValue;
        })
      );
    },
    [onListChange, list]
  );

  const onAddButtonClick = useCallback(() => {
    onListChange(
      produce(list, (draftList) => {
        draftList.push("");
      })
    );
  }, [list, onListChange]);

  const deleteTextField = useCallback(
    (index: number) => {
      onListChange(
        produce(list, (draftList) => {
          draftList.splice(index, 1);
        })
      );
    },
    [onListChange, list]
  );

  return (
    <section className={className}>
      {label ?? null}
      <Stack className={styles.list} tokens={{ childrenGap: 10 }}>
        {list.map((value, index) => (
          <TextFieldListItem
            key={index}
            index={index}
            parentJSONPointer={jsonPointer}
            getJSONPointer={getItemJSONPointer}
            textFieldProps={inputProps}
            value={value}
            onChange={onTextFieldChange}
            deleteItem={deleteTextField}
          />
        ))}
      </Stack>
      <Text className={styles.errorMessage}>
        {errorMessageProps ?? errorMessage}
      </Text>
      <ActionButton
        className={styles.addButton}
        theme={themes.actionButton}
        iconProps={{ iconName: "CirclePlus", className: styles.addButtonIcon }}
        onClick={onAddButtonClick}
      >
        <FormattedMessage id={addButtonLabelMessageID ?? "add"} />
      </ActionButton>
    </section>
  );
};

export default FormTextFieldList;
