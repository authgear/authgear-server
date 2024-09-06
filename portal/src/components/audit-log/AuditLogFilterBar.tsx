import React, { useCallback } from "react";
import cn from "classnames";
import styles from "./AuditLogFilterBar.module.css";
import { ISearchBoxProps, SearchBox } from "@fluentui/react";
import {
  DateRangeFilterDropdown,
  DateRangeFilterDropdownOptionKey,
} from "./DateRangeFilterDropdown";
import {
  ActivityTypeFilterDropdown,
  ActivityTypeFilterDropdownOptionKey,
} from "./ActivityTypeFilterDropdown";
import { AuditLogActivityType } from "../../graphql/adminapi/globalTypes.generated";

export interface AuditLogFilter {
  searchKeyword: string;
  activityType: ActivityTypeFilterDropdownOptionKey;
}

export interface AuditLogFilterBarPropsDateRange {
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
  availableActivityTypes: AuditLogActivityType[];
}

export const AuditLogFilterBar: React.VFC<AuditLogFilterBarProps> =
  function AuditLogFilterBar({
    className,
    filters,
    onFilterChange,
    searchBoxProps,
    dateRange,
    availableActivityTypes,
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
    const onChangeActivityType = useCallback(
      (newAT: ActivityTypeFilterDropdownOptionKey) => {
        onFilterChange((prev) => ({ ...prev, activityType: newAT }));
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
        <ActivityTypeFilterDropdown
          className={styles.activityTypeFilter}
          value={filters.activityType}
          onChange={onChangeActivityType}
          availableActivityTypes={availableActivityTypes}
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
