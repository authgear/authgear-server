import React, { useCallback, useContext, useMemo, useState } from "react";
import { Context as MessageContext } from "@oursky/react-messageformat";
import { IDropdownOption } from "@fluentui/react";
import { AuditLogActivityType } from "../../graphql/adminapi/globalTypes.generated";
import { SearchableDropdown } from "../common/SearchableDropdown";

export type AuditLogActivityTypeAll = "ALL";
export const ACTIVITY_TYPE_ALL: AuditLogActivityTypeAll = "ALL";

export type ActivityTypeFilterDropdownOptionKey =
  | AuditLogActivityType
  | AuditLogActivityTypeAll;

interface ActivityTypeFilterDropdownProps {
  className?: string;
  value: ActivityTypeFilterDropdownOptionKey;
  onChange: (newValue: ActivityTypeFilterDropdownOptionKey) => void;
  availableActivityTypes: AuditLogActivityType[];
}

export const ActivityTypeFilterDropdown: React.VFC<ActivityTypeFilterDropdownProps> =
  function ActivityTypeFilterDropdown({
    className,
    value,
    onChange,
    availableActivityTypes,
  }: ActivityTypeFilterDropdownProps) {
    const { renderToString } = useContext(MessageContext);
    const [searchValue, setSearchValue] = useState<string>("");

    const options = useMemo<IDropdownOption[]>(() => {
      return availableActivityTypes
        .map((key) => ({
          key,
          text: renderToString("AuditLogActivityType." + key),
        }))
        .filter((option) =>
          option.text.toLowerCase().includes(searchValue.toLowerCase())
        );
    }, [availableActivityTypes, renderToString, searchValue]);

    const onChangeOption = useCallback(
      (_e: unknown, option?: IDropdownOption) => {
        if (option == null) {
          return;
        }
        onChange(option.key as ActivityTypeFilterDropdownOptionKey);
      },
      [onChange]
    );

    const onClearFilter = useCallback(() => {
      setSearchValue("");
      onChange(ACTIVITY_TYPE_ALL);
    }, [onChange]);

    const selectedOption = useMemo(() => {
      const matched = options.find((option) => option.key === value);
      return matched ?? null;
    }, [options, value]);

    // Note extra layer of mapping here.
    //           ALL -> null              in SearchableDropdown
    // normal option -> normal option     in SearchableDropdown
    return (
      <SearchableDropdown
        className={className}
        placeholder={renderToString("AuditLogActivityType.ALL")}
        selectedItem={selectedOption}
        options={options}
        onChange={onChangeOption}
        searchValue={searchValue}
        onSearchValueChange={setSearchValue}
        onClear={onClearFilter}
      />
    );
  };
