import { IDropdownOption } from "@fluentui/react";
import React, { useCallback, useMemo, useState } from "react";
import { useDebounced } from "../../hook/useDebounced";
import { SearchableDropdown } from "../common/SearchableDropdown";

export function SearchableDropdownStory(): React.ReactElement {
  const [searchKeyword, setSearchKeyword] = useState("");
  const [selectedItem, setSelectedItem] = useState<IDropdownOption | null>(
    null
  );
  const allOptions = useMemo((): IDropdownOption[] => {
    return new Array(50).fill("").map<IDropdownOption>((_, idx) => {
      return {
        key: `${idx}`,
        text: `This is option ${idx + 1}`,
      };
    });
  }, []);

  const [deboundedKeyword, isDebouncing] = useDebounced(searchKeyword, 300);

  const options = useMemo(() => {
    return allOptions.filter((option) =>
      option.text.includes(deboundedKeyword.trim())
    );
  }, [allOptions, deboundedKeyword]);

  const onChange = useCallback((_: unknown, option?: IDropdownOption) => {
    if (!option) {
      setSelectedItem(null);
      return;
    }
    if (option.selected || option.selected === undefined) {
      setSelectedItem(option);
    } else {
      setSelectedItem(null);
    }
  }, []);

  const onClear = useCallback(() => {
    setSelectedItem(null);
  }, []);

  return (
    <SearchableDropdown
      placeholder="Select something"
      searchPlaceholder="Type here to search"
      optionsEmptyMessage="Nothing to show"
      isLoadingOptions={isDebouncing}
      options={options}
      searchValue={searchKeyword}
      onSearchValueChange={setSearchKeyword}
      selectedItem={selectedItem}
      onChange={onChange}
      onClear={onClear}
    />
  );
}
