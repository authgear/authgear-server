import React, { useCallback, useContext, useMemo } from "react";
import { Context as MessageContext } from "@oursky/react-messageformat";
import {
  Dropdown,
  IContextualMenuItem,
  IContextualMenuProps,
  IDropdownOption,
} from "@fluentui/react";
import CommandBarButton from "../../CommandBarButton";
import { AuditLogActivityType } from "../../graphql/adminapi/globalTypes.generated";
import { useScreenBreakpoint } from "../../hook/useScreenBreakpoint";

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

    const options = useMemo<IDropdownOption[]>(() => {
      const options: IDropdownOption[] = [
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

    const onChangeOption = useCallback(
      (_e: unknown, option?: IDropdownOption) => {
        if (option == null) {
          return;
        }
        onChange(option.key as ActivityTypeFilterDropdownOptionKey);
      },
      [onChange]
    );
    return (
      <Dropdown
        className={className}
        selectedKey={value}
        options={options}
        onChange={onChangeOption}
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
