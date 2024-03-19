import { useQuery } from "@apollo/client";
import { SearchBox } from "@fluentui/react";
import {
  FormattedMessage,
  Context as MessageContext,
} from "@oursky/react-messageformat";
import React, { useContext, useState, useCallback, useMemo } from "react";
import ShowError from "../../../ShowError";
import ShowLoading from "../../../ShowLoading";
import {
  RolesListQueryQuery,
  RolesListQueryQueryVariables,
  RolesListQueryDocument,
} from "../../../graphql/adminapi/query/rolesListQuery.generated";
import { UserQueryNodeFragment } from "../../../graphql/adminapi/query/userQuery.generated";
import { searchRoles } from "../../../model/role";
import { RolesEmptyView } from "../empty-view/RolesEmptyView";
import { UserRolesListItem, UserRolesList } from "./UserRolesList";
import PrimaryButton from "../../../PrimaryButton";
import cn from "classnames";

const pageSize = 10;

function UserDetailsScreenRoleListContainer({
  user,
  className,
}: {
  user: UserQueryNodeFragment;
  className?: string;
}): React.ReactElement {
  const { renderToString } = useContext(MessageContext);
  const {
    data: rolesQueryData,
    loading,
    error,
    refetch,
  } = useQuery<RolesListQueryQuery, RolesListQueryQueryVariables>(
    RolesListQueryDocument,
    {
      variables: {
        pageSize: 0,
        searchKeyword: "",
      },
      fetchPolicy: "network-only",
    }
  );

  const [searchKeyword, setSearchKeyword] = useState<string>("");
  const isSearch = searchKeyword !== "";
  const [offset, setOffset] = useState(0);

  const onChangeOffset = useCallback((offset) => {
    setOffset(offset);
  }, []);

  const onChangeSearchKeyword = useCallback(
    (e?: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      if (e === undefined) {
        return;
      }
      const value = e.currentTarget.value;
      setSearchKeyword(value);
      // Reset offset when search keyword was changed.
      setOffset(0);
    },
    []
  );
  const onClearSearchKeyword = useCallback(() => {
    setSearchKeyword("");
  }, []);

  const groupRoles = useMemo(() => {
    const groupsRolesTable: Record<string, UserRolesListItem> = {};
    user.groups?.edges?.forEach((edge) => {
      const group = edge?.node;
      if (group?.roles?.edges == null) {
        return;
      }
      group.roles.edges.forEach((roleEdge) => {
        const role = roleEdge?.node;
        if (role == null) {
          return;
        }
        if (role.key in groupsRolesTable) {
          groupsRolesTable[role.key].groups.push(group);
        } else {
          const roleWithGroups = {
            ...role,
            groups: [group],
          };
          groupsRolesTable[role.key] = roleWithGroups;
        }
      });
    });

    return Object.entries(groupsRolesTable).map(([_, value]) => value);
  }, [user.groups?.edges]);

  const userRoles: UserRolesListItem[] = useMemo(() => {
    return (
      user.roles?.edges?.flatMap((e) => {
        if (e?.node) {
          return [{ ...e.node, groups: [] }];
        }
        return [];
      }) ?? []
    );
  }, [user.roles?.edges]);

  const combinedRoles: UserRolesListItem[] = useMemo(() => {
    return [...groupRoles, ...userRoles];
  }, [groupRoles, userRoles]);

  const filteredCombinedRoles = useMemo(() => {
    if (isSearch) {
      return searchRoles(combinedRoles, searchKeyword);
    }

    return combinedRoles.slice(offset, offset + pageSize);
  }, [isSearch, combinedRoles, offset, searchKeyword]);

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  if (loading) {
    return <ShowLoading />;
  }

  const totalCount = rolesQueryData?.roles?.totalCount ?? 0;

  if (totalCount === 0) {
    return <RolesEmptyView />;
  }

  return (
    <>
      <section className={cn("flex flex-col h-full", className)}>
        <header className="flex flex-row items-center justify-between mb-8">
          <SearchBox
            className="max-w-[300px] min-w-0 flex-1 mr-2"
            placeholder={renderToString("search")}
            value={searchKeyword}
            onChange={onChangeSearchKeyword}
            onClear={onClearSearchKeyword}
          />
          <PrimaryButton
            text={<FormattedMessage id="UserDetailsScreen.roles.add" />}
          />
        </header>
        <UserRolesList
          className="flex-1 min-h-0"
          user={user}
          roles={filteredCombinedRoles}
          isSearch={isSearch}
          offset={offset}
          pageSize={pageSize}
          totalCount={combinedRoles.length}
          onChangeOffset={onChangeOffset}
        />
      </section>
    </>
  );
}

export default UserDetailsScreenRoleListContainer;
