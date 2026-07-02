import React, { useCallback, useContext } from "react";
import cn from "classnames";
import { Cross2Icon, MagnifyingGlassIcon } from "@radix-ui/react-icons";
import { TextField as RadixTextField } from "@radix-ui/themes";
import styles from "./AuditLogFilterBar.module.css";
import { ISearchBoxProps } from "@fluentui/react";
import { Context as MessageContext } from "../../intl";
import {
  AuditLogDateRangeFilterDropdown,
  AuditLogDateRangePresetKey,
} from "./AuditLogDateRangeFilterDropdown";
import {
  ActivityTypeFilterDropdown,
  ActivityTypeFilterDropdownOptionKey,
} from "./ActivityTypeFilterDropdown";
import { AuditLogActivityType } from "../../graphql/adminapi/globalTypes.generated";
import { RefreshButton } from "./RefreshButton";

export interface AuditLogFilter {
  searchKeyword: string;
  activityType: ActivityTypeFilterDropdownOptionKey;
}

export interface AuditLogFilterBarPropsDateRange {
  value: AuditLogDateRangePresetKey;
  onChange: (value: AuditLogDateRangePresetKey) => void;
  rangeFrom?: Date | null;
  rangeTo?: Date | null;
  onOpenCustomDateRangeDialog?: () => void;
}

interface AuditLogFilterBarProps {
  className?: string;
  filters: AuditLogFilter;
  onFilterChange: (fn: (prevValue: AuditLogFilter) => AuditLogFilter) => void;
  onRefresh: () => void;
  searchBoxProps?: ISearchBoxProps;
  dateRange: AuditLogFilterBarPropsDateRange;
  availableActivityTypes: AuditLogActivityType[];
  lastUpdatedAt: Date;
  wideActivityTypeDropdown?: boolean;
}

export const AuditLogFilterBar: React.VFC<AuditLogFilterBarProps> =
  function AuditLogFilterBar({
    className,
    filters,
    onFilterChange,
    onRefresh,
    searchBoxProps,
    dateRange,
    availableActivityTypes,
    lastUpdatedAt,
    wideActivityTypeDropdown = false,
  }) {
    const { renderToString } = useContext(MessageContext);

    const onChangeSearchKeyword = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.currentTarget.value;
        onFilterChange((prev) => ({
          ...prev,
          searchKeyword: value,
        }));
      },
      [onFilterChange]
    );
    const onClearSearchKeyword = useCallback(() => {
      onFilterChange((prev) => ({ ...prev, searchKeyword: "" }));
    }, [onFilterChange]);
    const onChangeActivityType = useCallback(
      (newAT: ActivityTypeFilterDropdownOptionKey) => {
        onFilterChange((prev) => ({ ...prev, activityType: newAT }));
      },
      [onFilterChange]
    );

    return (
      <div className={cn(styles.root, className)}>
        <div className={styles.filterContainer}>
          <AuditLogDateRangeFilterDropdown
            className={styles.dateRangeFilter}
            value={dateRange.value}
            onChange={dateRange.onChange}
            rangeFrom={dateRange.rangeFrom}
            rangeTo={dateRange.rangeTo}
            onOpenCustomDateRangeDialog={dateRange.onOpenCustomDateRangeDialog}
          />
          <ActivityTypeFilterDropdown
            className={styles.activityTypeFilter}
            value={filters.activityType}
            onChange={onChangeActivityType}
            availableActivityTypes={availableActivityTypes}
            wideContent={wideActivityTypeDropdown}
          />
          <RadixTextField.Root
            className={styles.searchBox}
            size="2"
            type="search"
            value={filters.searchKeyword}
            placeholder={searchBoxProps?.placeholder}
            onChange={onChangeSearchKeyword}
          >
            <RadixTextField.Slot side="left">
              <MagnifyingGlassIcon className={styles.searchIcon} />
            </RadixTextField.Slot>
            {filters.searchKeyword !== "" ? (
              <RadixTextField.Slot side="right">
                <button
                  type="button"
                  className={styles.searchClearButton}
                  aria-label={renderToString("AuditLogScreen.clear-search")}
                  onClick={onClearSearchKeyword}
                >
                  <Cross2Icon className={styles.searchClearIcon} />
                </button>
              </RadixTextField.Slot>
            ) : null}
          </RadixTextField.Root>
        </div>
        <div className={styles.filterActionContainer}>
          <RefreshButton onClick={onRefresh} lastUpdatedAt={lastUpdatedAt} />
        </div>
      </div>
    );
  };
