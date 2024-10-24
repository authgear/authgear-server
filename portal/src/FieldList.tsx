import cn from "classnames";
import React, {
  ComponentType,
  CSSProperties,
  useCallback,
  useMemo,
} from "react";
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
  listItemClassName?: string;
  listItemStyle?: CSSProperties;
  label?: React.ReactNode;
  parentJSONPointer: string | RegExp;
  fieldName: string;
  list: T[];
  onListItemChange: (list: T[], index: number, item: T) => void;
  onListItemAdd: (list: T[], item: T) => void;
  onListItemDelete: (list: T[], index: number, item: T) => void;
  makeDefaultItem: () => T;
  ListItemComponent: ComponentType<ListItemProps<T>>;
  addButtonLabelMessageID?: string;
  description?: string;
  descriptionPosition?: "top" | "bottom";
  addDisabled?: boolean;
  deleteDisabled?: boolean;
  minItem?: number;
  maxItem?: number;
}

const FieldList = function FieldList<T>(
  props: FieldListProps<T>
): React.ReactElement {
  const {
    className,
    listClassName,
    listItemClassName,
    listItemStyle,
    label,
    parentJSONPointer,
    fieldName,
    list,
    onListItemChange,
    onListItemAdd,
    onListItemDelete,
    ListItemComponent,
    makeDefaultItem,
    addButtonLabelMessageID,
    addDisabled,
    deleteDisabled,
    description,
    descriptionPosition = "bottom",
    minItem,
    maxItem,
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
      onListItemChange(newList, index, newValue);
    },
    [onListItemChange, list]
  );

  const onItemAdd = useCallback(() => {
    const newList = list.slice();
    const newItem = makeDefaultItem();
    newList.push(newItem);
    onListItemAdd(newList, newItem);
  }, [list, onListItemAdd, makeDefaultItem]);

  const onItemDelete = useCallback(
    (index: number) => {
      const item = list[index];
      const newList = list.slice();
      newList.splice(index, 1);
      onListItemDelete(newList, index, item);
    },
    [onListItemDelete, list]
  );

  const descriptionEl = useMemo(() => {
    if (description) {
      return (
        <Text
          block={true}
          className={cn(
            styles.description,
            descriptionPosition === "top" ? styles["description--top"] : null
          )}
        >
          {description}
        </Text>
      );
    }
    return null;
  }, [description, descriptionPosition]);

  const isMinItemReached = minItem != null && list.length <= minItem;
  const isMaxItemReached = maxItem != null && list.length >= maxItem;

  return (
    <div className={className}>
      {label ?? null}
      {descriptionPosition === "top" && descriptionEl ? descriptionEl : null}
      <div className={cn(styles.list, listClassName)}>
        {list.map((value, index) => (
          <FieldListItem
            className={listItemClassName}
            style={listItemStyle}
            key={index}
            index={index}
            value={value}
            onItemChange={onItemChange}
            onItemDelete={onItemDelete}
            ListItemComponent={ListItemComponent}
            deleteDisabled={deleteDisabled || isMinItemReached}
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
        disabled={addDisabled || isMaxItemReached}
      />
      {descriptionPosition === "bottom" && descriptionEl ? descriptionEl : null}
    </div>
  );
};

interface FieldListItemProps<T> {
  className?: string;
  style?: CSSProperties;
  index: number;
  value: T;
  onItemChange: (index: number, newValue: T) => void;
  onItemDelete: (index: number) => void;
  ListItemComponent: ComponentType<ListItemProps<T>>;
  deleteDisabled?: boolean;
}

function FieldListItem<T>(props: FieldListItemProps<T>) {
  const {
    className,
    style,
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
    <div className={cn(styles.listItem, className)} style={style}>
      <ListItemComponent index={index} value={value} onChange={onChange} />
      <IconButton
        className={cn(styles.deleteButton, deleteDisabled && "invisible")}
        onClick={onDelete}
        iconProps={{ iconName: "Delete" }}
        theme={themes.destructive}
        disabled={deleteDisabled}
      />
    </div>
  );
}

export default FieldList;
