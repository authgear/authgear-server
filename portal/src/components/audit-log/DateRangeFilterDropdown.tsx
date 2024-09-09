import React, { useCallback, useContext, useMemo } from "react";
import { Context as MessageContext } from "@oursky/react-messageformat";
import {
  Dropdown,
  IContextualMenuProps,
  IDropdownOption,
} from "@fluentui/react";
import CommandBarButton from "../../CommandBarButton";
import { useScreenBreakpoint } from "../../hook/useScreenBreakpoint";

export type DateRangeFilterDropdownOptionKey =
  | "allDateRange"
  | "customDateRange";

interface DateRangeFilterDropdownProps {
  className?: string;
  value: DateRangeFilterDropdownOptionKey;
  onClickAllDateRange: (
    e?: React.MouseEvent<unknown> | React.KeyboardEvent<unknown>
  ) => void;
  onClickCustomDateRange: (
    e?: React.MouseEvent<unknown> | React.KeyboardEvent<unknown>
  ) => void;
}

const DesktopDateRangeFilterDropdown: React.VFC<DateRangeFilterDropdownProps> =
  function DesktopDateRangeFilterDropdown({
    className,
    value,
    onClickAllDateRange,
    onClickCustomDateRange,
  }: DateRangeFilterDropdownProps) {
    const { renderToString } = useContext(MessageContext);
    const allDateRangeLabel = renderToString("AuditLogScreen.date-range.all");
    const customDateRangeLabel = renderToString(
      "AuditLogScreen.date-range.custom"
    );

    const placeholder = useMemo(() => {
      if (value === "customDateRange") {
        return renderToString("AuditLogScreen.date-range.custom");
      }

      return renderToString("AuditLogScreen.date-range.all");
    }, [renderToString, value]);

    const menuProps = useMemo<IContextualMenuProps>(() => {
      return {
        items: [
          {
            key: "allDateRange",
            text: allDateRangeLabel,
            onClick: onClickAllDateRange,
          },
          {
            key: "customDateRange",
            text: customDateRangeLabel,
            onClick: onClickCustomDateRange,
          },
        ],
      };
    }, [
      allDateRangeLabel,
      customDateRangeLabel,
      onClickAllDateRange,
      onClickCustomDateRange,
    ]);

    return (
      <CommandBarButton
        className={className}
        key="dateRange"
        iconProps={{ iconName: "Calendar" }}
        menuProps={menuProps}
        text={placeholder}
      />
    );
  };

const MobileDateRangeFilterDropdown: React.VFC<DateRangeFilterDropdownProps> =
  function MobileDateRangeFilterDropdown({
    className,
    value,
    onClickAllDateRange,
    onClickCustomDateRange,
  }: DateRangeFilterDropdownProps) {
    const { renderToString } = useContext(MessageContext);
    const allDateRangeLabel = renderToString("AuditLogScreen.date-range.all");
    const customDateRangeLabel = renderToString(
      "AuditLogScreen.date-range.custom"
    );
    const options = useMemo<IDropdownOption[]>(() => {
      return [
        {
          key: "allDateRange",
          text: allDateRangeLabel,
        },
        {
          key: "customDateRange",
          text: customDateRangeLabel,
        },
      ];
    }, [allDateRangeLabel, customDateRangeLabel]);

    const onChangeOption = useCallback(
      (_e: unknown, option?: IDropdownOption) => {
        if (option == null) {
          return;
        }
        switch (option.key) {
          case "allDateRange":
            onClickAllDateRange();
            break;
          case "customDateRange":
            onClickCustomDateRange();
            break;
          default:
            console.error("Unexpected option key: ", option.key);
            break;
        }
      },
      [onClickAllDateRange, onClickCustomDateRange]
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

export const DateRangeFilterDropdown: React.VFC<DateRangeFilterDropdownProps> =
  function DateRangeFilterDropdown(props: DateRangeFilterDropdownProps) {
    const screenBreakpoint = useScreenBreakpoint();
    switch (screenBreakpoint) {
      case "desktop": {
        return <DesktopDateRangeFilterDropdown {...props} />;
      }
      case "tablet":
      case "mobile":
        return <MobileDateRangeFilterDropdown {...props} />;
    }
  };
