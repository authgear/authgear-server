import React, { useCallback, useContext, useMemo, useState } from "react";
import { Context as MessageContext } from "@oursky/react-messageformat";
import {
  IContextualMenuItem,
  IContextualMenuProps,
  IDropdownOption,
} from "@fluentui/react";
import CommandBarButton from "../../CommandBarButton";
import { AuditLogActivityType } from "../../graphql/adminapi/globalTypes.generated";
import { useScreenBreakpoint } from "../../hook/useScreenBreakpoint";
import { SearchableDropdown } from "../common/SearchableDropdown";

export type AuditLogActivityTypeAll = "ALL";
export const ACTIVITY_TYPE_ALL: AuditLogActivityTypeAll = "ALL";

export type ActivityTypeFilterDropdownOptionKey =
  | AuditLogActivityType
  | AuditLogActivityTypeAll;

interface ActivityTypeDropdownOption {
  key: ActivityTypeFilterDropdownOptionKey;
  text: string;
}

interface ActivityTypeFilterDropdownProps {
  className?: string;
  value: ActivityTypeFilterDropdownOptionKey;
  onChange: (newValue: ActivityTypeFilterDropdownOptionKey) => void;
  availableActivityTypes: AuditLogActivityType[];
}

const DesktopActivityTypeFilterDropdown: React.VFC<ActivityTypeFilterDropdownProps> =
  function DesktopActivityTypeFilterDropdown({
    className,
    value,
    onChange,
    availableActivityTypes,
  }: ActivityTypeFilterDropdownProps) {
    const { renderToString } = useContext(MessageContext);

    const activityTypeOptions = useMemo<ActivityTypeDropdownOption[]>(() => {
      const options: ActivityTypeDropdownOption[] = [
        {
          key: ACTIVITY_TYPE_ALL,
          text: renderToString("AuditLogActivityType.ALL"),
        },
      ];
      for (const key of availableActivityTypes) {
        options.push({
          key: key,
          text: renderToString("AuditLogActivityType." + key),
        });
      }
      return options;
    }, [availableActivityTypes, renderToString]);

    const placeholder = useMemo(() => {
      return activityTypeOptions.find((option) => option.key === value)!.text;
    }, [activityTypeOptions, value]);

    const onClickOption = useCallback(
      (
        _event?:
          | React.MouseEvent<HTMLElement>
          | React.KeyboardEvent<HTMLElement>,
        item?: IContextualMenuItem
      ) => {
        onChange(item?.key as ActivityTypeFilterDropdownOptionKey);
      },
      [onChange]
    );

    const menuProps = useMemo<IContextualMenuProps>(() => {
      return {
        items: activityTypeOptions.map((option) => ({
          key: option.key,
          text: option.text,
          onClick: onClickOption,
        })),
      };
    }, [activityTypeOptions, onClickOption]);

    return (
      <CommandBarButton
        className={className}
        key="activityTypes"
        iconProps={{ iconName: "PC1" }}
        menuProps={menuProps}
        text={placeholder}
      />
    );
  };

const MobileActivityTypeFilterDropdown: React.VFC<ActivityTypeFilterDropdownProps> =
  function MobileActivityTypeFilterDropdown({
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

export const ActivityTypeFilterDropdown: React.VFC<ActivityTypeFilterDropdownProps> =
  function ActivityTypeFilterDropdown(props: ActivityTypeFilterDropdownProps) {
    const screenBreakpoint = useScreenBreakpoint();

    switch (screenBreakpoint) {
      case "desktop":
        return <DesktopActivityTypeFilterDropdown {...props} />;
      case "mobile":
      case "tablet":
        return <MobileActivityTypeFilterDropdown {...props} />;
    }
  };
