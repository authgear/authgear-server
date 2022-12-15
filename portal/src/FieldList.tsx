import cn from "classnames";
import React, { ComponentType, useCallback, useMemo } from "react";
import { IconButton, Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { useSystemConfig } from "./context/SystemConfigContext";
import { useFormField } from "./form";
import ErrorRenderer from "./ErrorRenderer";
import ActionButton from "./ActionButton";

import styles from "./FieldList.module.css";

export interface ListItemProps<T> {
  index: number;
  value: T;
  onChange: (value: T) => void;
}

export interface FieldListProps<T> {
  className?: string;
  listClassName?: string;
  label?: React.ReactNode;
  parentJSONPointer: string | RegExp;
  fieldName: string;
  list: T[];
  onListChange: (list: T[]) => void;
  makeDefaultItem: () => T;
  ListItemComponent: ComponentType<ListItemProps<T>>;
  addButtonLabelMessageID?: string;
  description?: string;
  addDisabled?: boolean;
  deleteDisabled?: boolean;
}

const FieldList = function FieldList<T>(
  props: FieldListProps<T>
): React.ReactElement {
  const {
    className,
    listClassName,
    label,
    parentJSONPointer,
    fieldName,
    list,
    onListChange,
    ListItemComponent,
    makeDefaultItem,
    addButtonLabelMessageID,
    addDisabled,
    deleteDisabled,
    description,
  } = props;

  const { themes } = useSystemConfig();

  const field = useMemo(
    () => ({
      parentJSONPointer,
      fieldName,
    }),
    [parentJSONPointer, fieldName]
  );
  const { errors } = useFormField(field);

  const onItemChange = useCallback(
    (index: number, newValue: T) => {
      const newList = list.slice();
      newList[index] = newValue;
      onListChange(newList);
    },
    [onListChange, list]
  );

  const onItemAdd = useCallback(() => {
    const newList = list.slice();
    newList.push(makeDefaultItem());
    onListChange(newList);
  }, [list, onListChange, makeDefaultItem]);

  const onItemDelete = useCallback(
    (index: number) => {
      const newList = list.slice();
      newList.splice(index, 1);
      onListChange(newList);
    },
    [onListChange, list]
  );

  return (
    <div className={className}>
      {label ?? null}
      <div className={cn(styles.list, listClassName)}>
        {list.map((value, index) => (
          <FieldListItem
            key={index}
            index={index}
            value={value}
            onItemChange={onItemChange}
            onItemDelete={onItemDelete}
            ListItemComponent={ListItemComponent}
            deleteDisabled={deleteDisabled}
          />
        ))}
      </div>
      <Text className={styles.errorMessage}>
        <ErrorRenderer errors={errors} />
      </Text>
      <ActionButton
        className={styles.addButton}
        theme={themes.actionButton}
        iconProps={{ iconName: "CirclePlus", className: styles.addButtonIcon }}
        onClick={onItemAdd}
        text={<FormattedMessage id={addButtonLabelMessageID ?? "add"} />}
        disabled={addDisabled}
      />
      {description ? (
        <Text block={true} className={styles.description}>
          {description}
        </Text>
      ) : null}
    </div>
  );
};

interface FieldListItemProps<T> {
  index: number;
  value: T;
  onItemChange: (index: number, newValue: T) => void;
  onItemDelete: (index: number) => void;
  ListItemComponent: ComponentType<ListItemProps<T>>;
  deleteDisabled?: boolean;
}

function FieldListItem<T>(props: FieldListItemProps<T>) {
  const {
    index,
    value,
    onItemChange,
    onItemDelete,
    ListItemComponent,
    deleteDisabled,
  } = props;
  const { themes } = useSystemConfig();

  const onChange = useCallback(
    (newValue: T) => onItemChange(index, newValue),
    [onItemChange, index]
  );
  const onDelete = useCallback(
    () => onItemDelete(index),
    [onItemDelete, index]
  );

  return (
    <div className={styles.listItem}>
      <ListItemComponent index={index} value={value} onChange={onChange} />
      <IconButton
        className={styles.deleteButton}
        onClick={onDelete}
        iconProps={{ iconName: "Delete" }}
        theme={themes.destructive}
        disabled={deleteDisabled}
      />
    </div>
  );
}

export default FieldList;
