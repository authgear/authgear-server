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
  GroupsListQueryQuery,
  GroupsListQueryQueryVariables,
  GroupsListQueryDocument,
} from "../../../graphql/adminapi/query/groupsListQuery.generated";
import { UserQueryNodeFragment } from "../../../graphql/adminapi/query/userQuery.generated";
import { searchGroups } from "../../../model/group";
import { GroupsEmptyView } from "../empty-view/GroupsEmptyView";
import { UserGroupsListItem, UserGroupsList } from "./UserGroupsList";
import PrimaryButton from "../../../PrimaryButton";
import cn from "classnames";

function UserDetailsScreenGroupListContainer({
  user,
  className,
}: {
  user: UserQueryNodeFragment;
  className?: string;
}): React.ReactElement {
  const { renderToString } = useContext(MessageContext);
  const {
    data: groupsQueryData,
    loading,
    error,
    refetch,
  } = useQuery<GroupsListQueryQuery, GroupsListQueryQueryVariables>(
    GroupsListQueryDocument,
    {
      variables: {
        pageSize: 0,
        searchKeyword: "",
      },
      fetchPolicy: "network-only",
    }
  );

  const [searchKeyword, setSearchKeyword] = useState<string>("");
  const onChangeSearchKeyword = useCallback(
    (e?: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      if (e === undefined) {
        return;
      }
      const value = e.currentTarget.value;
      setSearchKeyword(value);
    },
    []
  );
  const onClearSearchKeyword = useCallback(() => {
    setSearchKeyword("");
  }, []);

  const filteredUserGroups = useMemo(() => {
    const userGroups =
      user.groups?.edges?.flatMap<UserGroupsListItem>((edge) => {
        if (edge?.node != null) {
          return [edge.node];
        }
        return [];
      }) ?? [];
    return searchGroups(userGroups, searchKeyword);
  }, [user.groups?.edges, searchKeyword]);

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  if (loading) {
    return <ShowLoading />;
  }

  const totalCount = groupsQueryData?.groups?.totalCount ?? 0;

  if (totalCount === 0) {
    return <GroupsEmptyView />;
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
            text={<FormattedMessage id="UserDetailsScreen.groups.add" />}
          />
        </header>
        <UserGroupsList
          className="flex-1 min-h-0"
          user={user}
          groups={filteredUserGroups}
        />
      </section>
    </>
  );
}

export default UserDetailsScreenGroupListContainer;
