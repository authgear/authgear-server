import React, { useCallback, useMemo, useState } from "react";
import { IDropdownOption, ITag } from "@fluentui/react";

type TextFieldType = "integer" | "text";

function textFieldValidate(value?: string, type?: TextFieldType): boolean {
  switch (type) {
    case "integer":
      return /^[0-9]*$/.test(value ?? "");
    default:
      return true;
  }
}

export const useTextField = (
  initialValue: string,
  type?: TextFieldType
): { value: string; onChange: (_event: any, value?: string) => void } => {
  const [textFieldValue, setTextFieldValue] = React.useState(initialValue);
  const onChange = React.useCallback(
    (_event, value?: string) => {
      if (!textFieldValidate(value, type)) {
        return;
      }
      setTextFieldValue(value ?? "");
    },
    [setTextFieldValue, type]
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
  initialOption?: K,
  displayText?: (key: K) => string,
  hiddenSelections?: Set<K>
): {
  selectedKey?: K;
  options: IDropdownOption[];
  onChange: (_event: any, option?: IDropdownOption) => void;
} {
  const [selectedKey, setSelectedKey] = useState<K | undefined>(initialOption);
  const options = useMemo(
    () =>
      makeDropdownOptions(keyList, selectedKey, displayText, hiddenSelections),
    [selectedKey, displayText, keyList, hiddenSelections]
  );

  const onChange = useCallback((_event: any, option?: IDropdownOption) => {
    if (option == null) {
      return;
    }
    setSelectedKey(option.key.toString() as K);
  }, []);

  return {
    selectedKey,
    options,
    onChange,
  };
}
