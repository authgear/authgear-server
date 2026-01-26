import React, { useCallback, useContext, useMemo, useState } from "react";
import { SearchableDropdown } from "../common/SearchableDropdown";
import { Context as MessageContext } from "../../intl";
import { useQuery } from "@apollo/client";
import {
  GroupsListQueryDocument,
  GroupsListQueryQuery,
  GroupsListQueryQueryVariables,
} from "../../graphql/adminapi/query/groupsListQuery.generated";
import { IDropdownOption } from "@fluentui/react";
import { Group } from "../../graphql/adminapi/globalTypes.generated";

interface GroupsFilterDropdownProps {
  className?: string;
  value: GroupsFilterDropdownOption | null;
  onChange: (newValue: GroupsFilterDropdownOption | null) => void;
  onClear: () => void;
}

const MAX_OPTIONS = 100;

export interface GroupsFilterDropdownOption extends IDropdownOption {
  group: Pick<Group, "id" | "key" | "name">;
}

export const GroupsFilterDropdown: React.VFC<GroupsFilterDropdownProps> =
  function GroupsFilterDropdown({
    className,
    value,
    onChange: propsOnChange,
    onClear,
  }: GroupsFilterDropdownProps) {
    const { renderToString } = useContext(MessageContext);
    const [searchKeyword, setSearchKeyword] = useState("");

    const { data, loading } = useQuery<
      GroupsListQueryQuery,
      GroupsListQueryQueryVariables
    >(GroupsListQueryDocument, {
      variables: {
        pageSize: MAX_OPTIONS,
        searchKeyword: searchKeyword,
        cursor: null,
      },
      fetchPolicy: "network-only",
    });

    const options = useMemo<GroupsFilterDropdownOption[]>(() => {
      return (
        data?.groups?.edges?.flatMap((edge) => {
          const node = edge?.node;
          if (!node) {
            return [];
          }
          return [
            { ...node, group: node, text: node.name ?? node.key, key: node.id },
          ];
        }) ?? []
      );
    }, [data?.groups?.edges]);

    const onChange = useCallback(
      (_: unknown, option?: IDropdownOption) => {
        propsOnChange(
          (option as GroupsFilterDropdownOption | undefined) ?? null
        );
      },
      [propsOnChange]
    );

    return (
      <SearchableDropdown
        className={className}
        placeholder={renderToString("UsersScreen.filters.groups.placeholder")}
        isLoadingOptions={loading}
        options={options}
        searchValue={searchKeyword}
        onSearchValueChange={setSearchKeyword}
        selectedItem={value}
        onChange={onChange}
        onClear={onClear}
      />
    );
  };
