import React, { useCallback, useContext, useMemo } from "react";
import cn from "classnames";
import { Cross2Icon, MagnifyingGlassIcon } from "@radix-ui/react-icons";
import { Text, TextField as RadixTextField } from "@radix-ui/themes";
import styles from "./AuditLogFilterBar.module.css";
import { ISearchBoxProps } from "@fluentui/react";
import { Context as MessageContext, FormattedMessage } from "../../intl";
import {
  AuditLogDateRangeFilterDropdown,
  AuditLogDateRangePresetKey,
} from "./AuditLogDateRangeFilterDropdown";
import { ActivityTypeFilterDropdown } from "./ActivityTypeFilterDropdown";
import { AuditLogActivityType } from "../../graphql/adminapi/globalTypes.generated";
import { RefreshButton } from "./RefreshButton";

export interface AuditLogFilter {
  searchKeyword: string;
  activityTypes: AuditLogActivityType[];
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
  hideSearchBox?: boolean;
  dateRange: AuditLogFilterBarPropsDateRange;
  availableActivityTypes: AuditLogActivityType[];
  lastUpdatedAt: Date;
  wideActivityTypeDropdown?: boolean;
  trailingActions?: React.ReactNode;
}

export const AuditLogFilterBar: React.VFC<AuditLogFilterBarProps> =
  function AuditLogFilterBar({
    className,
    filters,
    onFilterChange,
    onRefresh,
    searchBoxProps,
    hideSearchBox = false,
    dateRange,
    availableActivityTypes,
    lastUpdatedAt,
    wideActivityTypeDropdown = false,
    trailingActions,
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
    const onChangeActivityTypes = useCallback(
      (newActivityTypes: AuditLogActivityType[]) => {
        onFilterChange((prev) => ({
          ...prev,
          activityTypes: newActivityTypes,
        }));
      },
      [onFilterChange]
    );

    const selectedActivityTypeSummary = useMemo(() => {
      if (filters.activityTypes.length === 0) {
        return null;
      }
      return filters.activityTypes
        .map((activityType) =>
          renderToString("AuditLogActivityType." + activityType)
        )
        .sort((a, b) => a.localeCompare(b))
        .join(", ");
    }, [filters.activityTypes, renderToString]);

    return (
      <div className={cn(styles.root, className)}>
        <div className={styles.filtersRow}>
          <div className={styles.filterContainer}>
            <AuditLogDateRangeFilterDropdown
              className={styles.dateRangeFilter}
              value={dateRange.value}
              onChange={dateRange.onChange}
              rangeFrom={dateRange.rangeFrom}
              rangeTo={dateRange.rangeTo}
              onOpenCustomDateRangeDialog={
                dateRange.onOpenCustomDateRangeDialog
              }
            />
            <ActivityTypeFilterDropdown
              className={styles.activityTypeFilter}
              value={filters.activityTypes}
              onChange={onChangeActivityTypes}
              availableActivityTypes={availableActivityTypes}
              wideContent={wideActivityTypeDropdown}
            />
            {hideSearchBox ? null : (
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
            )}
          </div>
          <div className={styles.filterActionContainer}>
            <RefreshButton onClick={onRefresh} lastUpdatedAt={lastUpdatedAt} />
            {trailingActions}
          </div>
        </div>
        {selectedActivityTypeSummary != null ? (
          <Text
            as="p"
            size="1"
            color="gray"
            className={styles.selectedActivityTypesSummary}
            title={selectedActivityTypeSummary}
          >
            <FormattedMessage
              id="AuditLogScreen.activity-types-selected-summary"
              values={{ types: selectedActivityTypeSummary }}
            />
          </Text>
        ) : null}
      </div>
    );
  };
