import React, { useCallback } from "react";
import { Tag } from "../CustomTagPicker";
import { deduplicate } from "../util/array";

export function useTextField(onChange: (value: string) => void): {
  onChange: (_event: any, value?: string) => void;
} {
  const onTextFieldChange = useCallback(
    (_event, value?: string) => {
      onChange(value ?? "");
    },
    [onChange]
  );
  return {
    onChange: onTextFieldChange,
  };
}

export function useCheckbox(onChange: (checked: boolean) => void): {
  onChange: (_event: any, checked?: boolean) => void;
} {
  const onCheckboxChange = useCallback(
    (_event, checked?: boolean) => {
      if (checked == null) {
        return;
      }
      onChange(checked);
    },
    [onChange]
  );

  return { onChange: onCheckboxChange };
}

export const useTagPickerWithNewTags = (
  list: string[],
  onListChange: (list: string[]) => void
): {
  selectedItems: Tag[];
  onChange: (items?: Tag[]) => void;
  onResolveSuggestions: (filterText: string, _tagList?: Tag[]) => Tag[];
  onAdd: (value: string) => void;
} => {
  const onChange = React.useCallback(
    (items?: Tag[]) => {
      if (items == null) {
        return;
      }
      const listItems = deduplicate(items.map((item) => item.name)).filter(
        Boolean
      );
      onListChange(listItems);
    },
    [onListChange]
  );

  const onAdd = React.useCallback(
    (value: string) => {
      const listItems = deduplicate([...list, value]).filter(Boolean);
      onListChange(listItems);
    },
    [onListChange, list]
  );

  const selectedItems = React.useMemo(
    () =>
      list.map((text) => ({
        key: text,
        name: text,
      })),
    [list]
  );

  const onResolveSuggestions = React.useCallback(
    (filterText: string, _tagList?: Tag[]): Tag[] => {
      return [{ key: filterText, name: filterText }];
    },
    []
  );

  return {
    selectedItems,
    onChange,
    onResolveSuggestions,
    onAdd,
  };
};
