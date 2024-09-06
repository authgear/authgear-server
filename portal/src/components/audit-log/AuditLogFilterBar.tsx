import React, { useCallback } from "react";
import cn from "classnames";
import styles from "./AuditLogFilterBar.module.css";
import { ISearchBoxProps, SearchBox } from "@fluentui/react";
import {
  DateRangeFilterDropdown,
  DateRangeFilterDropdownOptionKey,
} from "./DateRangeFilterDropdown";

export interface AuditLogFilter {
  searchKeyword: string;
  // TODO: add below
  // activityType: ActivityTypeDropdownOption | null;
}

interface AuditLogFilterBarPropsDateRange {
  value: DateRangeFilterDropdownOptionKey;
  onClickAllDateRange: (
    e?: React.MouseEvent<unknown> | React.KeyboardEvent<unknown>
  ) => void;
  onClickCustomDateRange: (
    e?: React.MouseEvent<unknown> | React.KeyboardEvent<unknown>
  ) => void;
}

interface AuditLogFilterBarProps {
  className?: string;
  filters: AuditLogFilter;
  onFilterChange: (fn: (prevValue: AuditLogFilter) => AuditLogFilter) => void;
  searchBoxProps?: ISearchBoxProps;
  dateRange: AuditLogFilterBarPropsDateRange;
}

export const AuditLogFilterBar: React.VFC<AuditLogFilterBarProps> =
  function AuditLogFilterBar({
    className,
    filters,
    onFilterChange,
    searchBoxProps,
    dateRange,
  }) {
    const onChangeSearchKeyword = useCallback(
      (e?: React.ChangeEvent<HTMLInputElement>) => {
        if (e === undefined) {
          return;
        }
        onFilterChange((prev) => ({
          ...prev,
          searchKeyword: e.currentTarget.value,
        }));
      },
      [onFilterChange]
    );
    const onClearSearchKeyword = useCallback(() => {
      onFilterChange((prev) => ({ ...prev, searchKeyword: "" }));
    }, [onFilterChange]);

    return (
      <div className={cn(styles.root, className)}>
        <DateRangeFilterDropdown
          className={styles.dateRangeFilter}
          value={dateRange.value}
          onClickAllDateRange={dateRange.onClickAllDateRange}
          onClickCustomDateRange={dateRange.onClickCustomDateRange}
        />
        <SearchBox
          className={styles.searchBox}
          value={filters.searchKeyword}
          onChange={onChangeSearchKeyword}
          onClear={onClearSearchKeyword}
          {...searchBoxProps}
        />
      </div>
    );
  };
