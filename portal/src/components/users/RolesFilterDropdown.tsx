import React, { useCallback, useContext, useMemo, useState } from "react";
import { SearchableDropdown } from "../common/SearchableDropdown";
import { Context as MessageContext } from "@oursky/react-messageformat";
import { useQuery } from "@apollo/client";
import { IDropdownOption } from "@fluentui/react";
import { Role } from "../../graphql/adminapi/globalTypes.generated";
import {
  RolesListQueryDocument,
  RolesListQueryQuery,
  RolesListQueryQueryVariables,
} from "../../graphql/adminapi/query/rolesListQuery.generated";

interface RolesFilterDropdownProps {
  className?: string;
  value: RolesFilterDropdownOption | null;
  onChange: (newValue: RolesFilterDropdownOption | null) => void;
  onClear: () => void;
}

const MAX_OPTIONS = 100;

export interface RolesFilterDropdownOption extends IDropdownOption {
  role: Pick<Role, "id" | "key" | "name">;
}

export const RolesFilterDropdown: React.VFC<RolesFilterDropdownProps> =
  function RolesFilterDropdown({
    className,
    value,
    onChange: propsOnChange,
    onClear,
  }: RolesFilterDropdownProps) {
    const { renderToString } = useContext(MessageContext);
    const [searchKeyword, setSearchKeyword] = useState("");

    const { data, loading } = useQuery<
      RolesListQueryQuery,
      RolesListQueryQueryVariables
    >(RolesListQueryDocument, {
      variables: {
        pageSize: MAX_OPTIONS,
        searchKeyword: searchKeyword,
        cursor: null,
      },
      fetchPolicy: "network-only",
    });

    const options = useMemo<RolesFilterDropdownOption[]>(() => {
      return (
        data?.roles?.edges?.flatMap((edge) => {
          const node = edge?.node;
          if (!node) {
            return [];
          }
          return [
            { ...node, role: node, text: node.name ?? node.key, key: node.id },
          ];
        }) ?? []
      );
    }, [data?.roles?.edges]);

    const onChange = useCallback(
      (_: unknown, option?: IDropdownOption) => {
        propsOnChange(
          (option as RolesFilterDropdownOption | undefined) ?? null
        );
      },
      [propsOnChange]
    );

    return (
      <SearchableDropdown
        className={className}
        placeholder={renderToString("UsersScreen.filters.roles.placeholder")}
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
