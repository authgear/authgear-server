import React, { useCallback, useContext, useMemo, useState } from "react";
import { Context as MessageContext } from "../../intl";
import { useQuery } from "@apollo/client";
import {
  GroupsListQueryDocument,
  GroupsListQueryQuery,
  GroupsListQueryQueryVariables,
} from "../../graphql/adminapi/query/groupsListQuery.generated";
import { Group } from "../../graphql/adminapi/globalTypes.generated";
import {
  UsersFilterDropdown,
  UsersFilterDropdownOption,
} from "./UsersFilterDropdown";

interface GroupsFilterDropdownProps {
  className?: string;
  value: GroupsFilterDropdownOption | null;
  onChange: (newValue: GroupsFilterDropdownOption | null) => void;
  onClear: () => void;
}

const MAX_OPTIONS = 100;

export interface GroupsFilterDropdownOption extends UsersFilterDropdownOption {
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
      (option: GroupsFilterDropdownOption) => {
        propsOnChange(option);
      },
      [propsOnChange]
    );

    return (
      <UsersFilterDropdown
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
