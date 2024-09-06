import React, { useContext, useMemo } from "react";
import { Context as MessageContext } from "@oursky/react-messageformat";
import { IContextualMenuProps } from "@fluentui/react";
import CommandBarButton from "../../CommandBarButton";

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

export const DateRangeFilterDropdown: React.VFC<DateRangeFilterDropdownProps> =
  function DateRangeFilterDropdown({
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
