import React, { useCallback, useMemo } from "react";
import { IDropdownOption, ITag } from "@fluentui/react";

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
  selectedItems: ITag[];
  onChange: (items?: ITag[]) => void;
  onResolveSuggestions: (filterText: string, _tagList?: ITag[]) => ITag[];
} => {
  const onChange = React.useCallback(
    (items?: ITag[]) => {
      if (items == null) {
        return;
      }
      const listItems = items.map((item) => item.name);
      onListChange(listItems);
    },
    [onListChange]
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
    (filterText: string, _tagList?: ITag[]): ITag[] => {
      return [{ key: filterText, name: filterText }];
    },
    []
  );

  return {
    selectedItems,
    onChange,
    onResolveSuggestions,
  };
};

export function makeDropdownOptions<K extends string>(
  keyList: K[],
  selectedKey?: K,
  displayText?: (key: K) => string,
  hiddenSelections?: Set<K>
): IDropdownOption[] {
  return keyList.map((key) => ({
    key,
    text: displayText != null ? displayText(key) : key,
    isSelected: selectedKey === key,
    hidden: hiddenSelections?.has(key),
  }));
}

export function useDropdown<K extends string>(
  keyList: K[],
  onChange: (option: K) => void,
  selectedKey?: K,
  displayText?: (key: K) => string,
  hiddenSelections?: Set<K>
): {
  options: IDropdownOption[];
  onChange: (_event: any, option?: IDropdownOption) => void;
} {
  const options = useMemo(
    () =>
      makeDropdownOptions(keyList, selectedKey, displayText, hiddenSelections),
    [selectedKey, displayText, keyList, hiddenSelections]
  );

  const onSelectionChange = useCallback(
    (_event: any, option?: IDropdownOption) => {
      if (option == null) {
        return;
      }
      onChange(option.key.toString() as K);
    },
    [onChange]
  );

  return {
    options,
    onChange: onSelectionChange,
  };
}
