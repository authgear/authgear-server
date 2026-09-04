import React, { useCallback, useContext } from "react";
import cn from "classnames";
import styles from "./UsersFilterBar.module.css";
import { TextField as RadixTextField } from "@radix-ui/themes";
import { Cross2Icon, MagnifyingGlassIcon } from "@radix-ui/react-icons";
import { Context as MessageContext } from "../../intl";
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
          <RadixTextField.Root
            className={styles.searchBox}
            size="2"
            type="search"
            placeholder={renderToString("search")}
            disabled={isSearchDisabled}
            value={filters.searchKeyword}
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
                  aria-label={renderToString("APIResourcesScreen.clear-search")}
                  onClick={onClearSearchKeyword}
                >
                  <Cross2Icon className={styles.searchClearIcon} />
                </button>
              </RadixTextField.Slot>
            ) : null}
          </RadixTextField.Root>
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
