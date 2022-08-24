import React, { useCallback, useMemo } from "react";
import cn from "classnames";
import { ActionButton, IconButton, Stack, Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { useSystemConfig } from "./context/SystemConfigContext";
import { useFormField } from "./form";
import ErrorRenderer from "./ErrorRenderer";

import styles from "./FieldList.module.css";

type RenderFieldListItem<T> = (
  index: number,
  value: T,
  onChange: (value: T) => void
) => React.ReactElement;

interface FieldListProps<T> {
  className?: string;
  label?: React.ReactNode;
  parentJSONPointer: string | RegExp;
  fieldName: string;
  list: T[];
  onListChange: (list: T[]) => void;
  makeDefaultItem: () => T;
  renderListItem: RenderFieldListItem<T>;
  addButtonLabelMessageID?: string;
  addDisabled?: boolean;
  description?: string;
}

const FieldList = function FieldList<T>(
  props: FieldListProps<T>
): React.ReactElement {
  const {
    className,
    label,
    parentJSONPointer,
    fieldName,
    list,
    onListChange,
    renderListItem,
    makeDefaultItem,
    addButtonLabelMessageID,
    addDisabled,
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
    <section className={className}>
      {label ?? null}
      <Stack className={styles.list} tokens={{ childrenGap: 10 }}>
        {list.map((value, index) => (
          <FieldListItem
            key={index}
            index={index}
            value={value}
            onItemChange={onItemChange}
            onItemDelete={onItemDelete}
            renderListItem={renderListItem}
          />
        ))}
      </Stack>
      <Text className={styles.errorMessage}>
        <ErrorRenderer errors={errors} />
      </Text>
      <ActionButton
        className={cn(styles.addButton, {
          [styles.readOnly]: addDisabled,
        })}
        theme={themes.actionButton}
        iconProps={{ iconName: "CirclePlus", className: styles.addButtonIcon }}
        onClick={onItemAdd}
      >
        <FormattedMessage id={addButtonLabelMessageID ?? "add"} />
      </ActionButton>
      {description ? (
        <Text block={true} className={styles.description}>
          {description}
        </Text>
      ) : null}
    </section>
  );
};

interface FieldListItemProps<T> {
  index: number;
  value: T;
  onItemChange: (index: number, newValue: T) => void;
  onItemDelete: (index: number) => void;
  renderListItem: RenderFieldListItem<T>;
}

function FieldListItem<T>(props: FieldListItemProps<T>) {
  const { index, value, onItemChange, onItemDelete, renderListItem } = props;
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
      {renderListItem(index, value, onChange)}
      <IconButton
        className={styles.deleteButton}
        onClick={onDelete}
        iconProps={{ iconName: "Delete" }}
        theme={themes.destructive}
      />
    </div>
  );
}

export default FieldList;
