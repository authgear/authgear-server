import React, { useCallback, useMemo, useState } from "react";
import { IDropdownOption, ITag } from "@fluentui/react";

export const useTextField = (
  initialValue: string
): { value: string; onChange: (_event: any, value?: string) => void } => {
  const [textFieldValue, setTextFieldValue] = React.useState(initialValue);
  const onChange = React.useCallback(
    (_event, value?: string) => {
      setTextFieldValue(value ?? "");
    },
    [setTextFieldValue]
  );
  return {
    value: textFieldValue,
    onChange,
  };
};

export const useCheckbox = (
  initialValue: boolean
): { value: boolean; onChange: (_event: any, value?: boolean) => void } => {
  const [checked, setChecked] = React.useState(initialValue);
  const onChange = React.useCallback(
    (_event, value?: boolean) => {
      setChecked(!!value);
    },
    [setChecked]
  );
  return {
    value: checked,
    onChange,
  };
};

export const useTagPickerWithNewTags = (
  initialList: string[],
  suggestionList?: ITag[]
): {
  list: string[];
  defaultSelectedItems: ITag[];
  onChange: (items?: ITag[]) => void;
  onResolveSuggestions: (filterText: string, _tagList?: ITag[]) => ITag[];
} => {
  const [list, setList] = React.useState(initialList);

  const onChange = React.useCallback((items?: ITag[]) => {
    if (items == null) {
      return;
    }
    const listItems = items.map((item) => item.name);
    setList(listItems);
  }, []);

  const onResolveSuggestions = React.useCallback(
    (filterText: string, _tagList?: ITag[]): ITag[] => {
      if (!suggestionList) {
        return [{ key: filterText, name: filterText }];
      }
      const matches = suggestionList.filter((tag) =>
        tag.name.toLowerCase().includes(filterText)
      );
      return matches.concat({ key: filterText, name: filterText });
    },
    [suggestionList]
  );

  const defaultSelectedItems = React.useMemo(
    () =>
      initialList.map((text) => ({
        key: text,
        name: text,
      })),
    [initialList]
  );

  return {
    list,
    defaultSelectedItems,
    onChange,
    onResolveSuggestions,
  };
};

export function makeDropdownOptions(
  keyList: string[],
  selectedKey?: string,
  displayText?: (key: string) => string
): IDropdownOption[] {
  return keyList.map((key) => ({
    key,
    text: displayText != null ? displayText(key) : key,
    isSelected: selectedKey === key,
  }));
}

export const useDropdown = (
  keyList: string[],
  initialOption?: string,
  displayText?: (key: string) => string
): {
  selectedKey?: string;
  options: IDropdownOption[];
  onChange: (_event: any, option?: IDropdownOption) => void;
} => {
  const [selectedKey, setSelectedKey] = useState<string | undefined>(
    initialOption
  );
  const options = useMemo(
    () => makeDropdownOptions(keyList, selectedKey, displayText),
    [selectedKey, displayText, keyList]
  );

  const onChange = useCallback((_event: any, option?: IDropdownOption) => {
    if (option == null) {
      return;
    }
    setSelectedKey(option.key.toString());
  }, []);

  return {
    selectedKey,
    options,
    onChange,
  };
};
