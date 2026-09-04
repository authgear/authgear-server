import React, { useCallback, useMemo } from "react";
import cn from "classnames";
import { IconButton as RadixIconButton, Text } from "@radix-ui/themes";
import { TrashIcon } from "@radix-ui/react-icons";
import { FormattedMessage } from "../../../intl";
import { useFormField } from "../../../form";
import { joinParentChild } from "../../../util/jsonpointer";
import ErrorRenderer from "../../../ErrorRenderer";
import { TextField } from "../TextField/TextField";
import { SecondaryButton } from "../Button/SecondaryButton/SecondaryButton";
import styles from "./TextFieldList.module.css";

export interface TextFieldListProps {
  className?: string;
  label?: React.ReactNode;
  description?: React.ReactNode;
  parentJSONPointer: string | RegExp;
  fieldName: string;
  placeholder?: string;
  list: string[];
  onListItemAdd: (list: string[], item: string) => void;
  onListItemChange: (list: string[], index: number, item: string) => void;
  onListItemDelete: (list: string[], index: number, item: string) => void;
  addButtonLabelMessageID?: string;
  deleteButtonAriaLabel?: string;
  disabled?: boolean;
  minItem?: number;
  maxItem?: number;
}

interface TextFieldListItemProps {
  index: number;
  itemsJSONPointer: string | RegExp;
  placeholder?: string;
  value: string;
  disabled?: boolean;
  canDelete: boolean;
  deleteButtonAriaLabel?: string;
  onItemChange: (index: number, value: string) => void;
  onItemDelete: (index: number) => void;
}

function TextFieldListItem({
  index,
  itemsJSONPointer,
  placeholder,
  value,
  disabled,
  canDelete,
  deleteButtonAriaLabel,
  onItemChange,
  onItemDelete,
}: TextFieldListItemProps): React.ReactElement {
  const onChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      onItemChange(index, e.currentTarget.value);
    },
    [index, onItemChange]
  );
  const onDeleteClick = useCallback(() => {
    onItemDelete(index);
  }, [index, onItemDelete]);

  return (
    <div className={styles.row}>
      <div className={styles.rowField}>
        <TextField
          size="2"
          parentJSONPointer={itemsJSONPointer}
          fieldName={index.toString(10)}
          placeholder={placeholder}
          value={value}
          onChange={onChange}
          disabled={disabled}
        />
      </div>
      {canDelete ? (
        <RadixIconButton
          className={styles.deleteButton}
          variant="ghost"
          color="gray"
          size="2"
          type="button"
          aria-label={deleteButtonAriaLabel}
          disabled={disabled}
          onClick={onDeleteClick}
        >
          <TrashIcon width="1rem" height="1rem" />
        </RadixIconButton>
      ) : null}
    </div>
  );
}

// The v2 counterpart of FormTextFieldList: a growable list of single-line
// text fields with per-item form error binding (items bind to
// <parent>/<fieldName>/<index>), a list-level error message, and add/delete
// controls.
export function TextFieldList(props: TextFieldListProps): React.ReactElement {
  const {
    className,
    label,
    description,
    parentJSONPointer,
    fieldName,
    placeholder,
    list: propList,
    onListItemAdd,
    onListItemChange,
    onListItemDelete,
    addButtonLabelMessageID,
    deleteButtonAriaLabel,
    disabled,
    minItem,
    maxItem,
  } = props;

  const field = useMemo(
    () => ({ parentJSONPointer, fieldName }),
    [parentJSONPointer, fieldName]
  );
  const { errors } = useFormField(field);

  const itemsJSONPointer = useMemo(
    () => joinParentChild(parentJSONPointer, fieldName),
    [parentJSONPointer, fieldName]
  );

  const list = useMemo(() => {
    // If the number of items is less than minItem, fill with empty items.
    if (minItem == null || minItem === 0) {
      return propList;
    }
    if (propList.length === 0) {
      return new Array(minItem).fill("") as string[];
    }
    return propList;
  }, [minItem, propList]);

  const canDelete = minItem == null || list.length > minItem;
  const canAdd = maxItem == null || list.length < maxItem;

  const onItemChange = useCallback(
    (index: number, value: string) => {
      onListItemChange(list, index, value);
    },
    [list, onListItemChange]
  );

  const onItemDelete = useCallback(
    (index: number) => {
      onListItemDelete(list, index, list[index]);
    },
    [list, onListItemDelete]
  );

  const onAddClick = useCallback(() => {
    onListItemAdd(list, "");
  }, [list, onListItemAdd]);

  return (
    <div className={cn(styles.root, className)}>
      {label != null ? (
        <Text as="p" size="2" weight="medium" className={styles.label}>
          {label}
        </Text>
      ) : null}
      <div className={styles.list}>
        {list.map((value, index) => (
          <TextFieldListItem
            key={index}
            index={index}
            itemsJSONPointer={itemsJSONPointer}
            placeholder={placeholder}
            value={value}
            disabled={disabled}
            canDelete={canDelete}
            deleteButtonAriaLabel={deleteButtonAriaLabel}
            onItemChange={onItemChange}
            onItemDelete={onItemDelete}
          />
        ))}
      </div>
      {errors.length > 0 ? (
        <Text as="p" size="1" color="red" className={styles.errors}>
          <ErrorRenderer errors={errors} />
        </Text>
      ) : null}
      {canAdd ? (
        <span className={styles.addButton}>
          <SecondaryButton
            size="2"
            disabled={disabled}
            onClick={onAddClick}
            text={<FormattedMessage id={addButtonLabelMessageID ?? "add"} />}
          />
        </span>
      ) : null}
      {description != null ? (
        <Text as="p" size="1" className={styles.description}>
          {description}
        </Text>
      ) : null}
    </div>
  );
}
