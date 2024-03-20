import React, { useCallback, useContext } from "react";
import cn from "classnames";
import styles from "./UsersFilterBar.module.css";
import { SearchBox } from "@fluentui/react";
import { Context as MessageContext } from "@oursky/react-messageformat";
import {
  GroupsFilterDropdown,
  GroupsFilterDropdownOption,
} from "./GroupsFilterDropdown";
import {
  RolesFilterDropdown,
  RolesFilterDropdownOption,
} from "./RolesFilterDropdown";

export interface UsersFilter {
  searchKeyword: string;
  group: GroupsFilterDropdownOption | null;
  role: RolesFilterDropdownOption | null;
}

interface UsersFilterBarProps {
  className?: string;
  showSearchBar: boolean;
  isSearchDisabled?: boolean;
  showRoleFilter: boolean;
  showGroupFilter: boolean;
  filters: UsersFilter;
  onFilterChange: (fn: (prevValue: UsersFilter) => UsersFilter) => void;
}

export const UsersFilterBar: React.VFC<UsersFilterBarProps> =
  function UsersFilterBar({
    className,
    showSearchBar,
    isSearchDisabled,
    showGroupFilter,
    showRoleFilter,
    filters,
    onFilterChange,
  }) {
    const { renderToString } = useContext(MessageContext);

    const onChangeSearchKeyword = useCallback(
      (_: unknown, newValue?: string) => {
        onFilterChange((prev) => ({ ...prev, searchKeyword: newValue ?? "" }));
      },
      [onFilterChange]
    );
    const onClearSearchKeyword = useCallback(() => {
      onFilterChange((prev) => ({ ...prev, searchKeyword: "" }));
    }, [onFilterChange]);

    const onGroupChange = useCallback(
      (newValue: GroupsFilterDropdownOption | null) => {
        onFilterChange((prev) => ({ ...prev, group: newValue }));
      },
      [onFilterChange]
    );

    const onGroupClear = useCallback(() => {
      onFilterChange((prev) => ({ ...prev, group: null }));
    }, [onFilterChange]);

    const onRoleChange = useCallback(
      (newValue: RolesFilterDropdownOption | null) => {
        onFilterChange((prev) => ({ ...prev, role: newValue }));
      },
      [onFilterChange]
    );

    const onRoleClear = useCallback(() => {
      onFilterChange((prev) => ({ ...prev, role: null }));
    }, [onFilterChange]);

    return (
      <div className={cn(styles.root, className)}>
        {showSearchBar ? (
          <SearchBox
            className={styles.searchBox}
            placeholder={renderToString("search")}
            disabled={isSearchDisabled}
            value={filters.searchKeyword}
            onChange={onChangeSearchKeyword}
            onClear={onClearSearchKeyword}
          />
        ) : null}
        <div className={styles.filterContainer}>
          {showRoleFilter ? (
            <RolesFilterDropdown
              className={styles.filter}
              value={filters.role}
              onChange={onRoleChange}
              onClear={onRoleClear}
            />
          ) : null}
          {showGroupFilter ? (
            <GroupsFilterDropdown
              className={styles.filter}
              value={filters.group}
              onChange={onGroupChange}
              onClear={onGroupClear}
            />
          ) : null}
        </div>
      </div>
    );
  };
