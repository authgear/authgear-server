import React, { useCallback } from "react";
import cn from "classnames";
import styles from "./AuditLogFilterBar.module.css";
import { ISearchBoxProps, SearchBox } from "@fluentui/react";

export interface AuditLogFilter {
  searchKeyword: string;
  // TODO: add below
  // dateRange: DateRangeDropdownOption | null;
  // activityType: ActivityTypeDropdownOption | null;
}

interface AuditLogFilterBarProps {
  className?: string;
  filters: AuditLogFilter;
  onFilterChange: (fn: (prevValue: AuditLogFilter) => AuditLogFilter) => void;
  searchBoxProps?: ISearchBoxProps;
}

export const AuditLogFilterBar: React.VFC<AuditLogFilterBarProps> =
  function AuditLogFilterBar({
    className,
    filters,
    onFilterChange,
    searchBoxProps,
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
