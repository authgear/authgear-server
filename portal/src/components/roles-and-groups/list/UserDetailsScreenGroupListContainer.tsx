import { useQuery } from "@apollo/client";
import { FormattedMessage, Context as MessageContext } from "../../../intl";
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
import cn from "classnames";
import { AddUserGroupsDialog } from "../dialog/AddUserGroupsDialog";
import { TextField, TextFieldIcon } from "../../v2/TextField/TextField";
import { PrimaryButton } from "../../v2/Button/PrimaryButton/PrimaryButton";
import { Callout as RadixCallout } from "@radix-ui/themes";
import { InfoCircledIcon, PlusIcon } from "@radix-ui/react-icons";

const pageSize = 10;

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
  const isSearch = searchKeyword !== "";
  const [offset, setOffset] = useState(0);

  const onChangeOffset = useCallback((offset) => {
    setOffset(offset);
  }, []);

  const onChangeSearchKeyword = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.currentTarget.value;
      setSearchKeyword(value);
      // Reset offset when search keyword was changed.
      setOffset(0);
    },
    []
  );
  const [isAddRoleDialogHidden, setIsAddRoleDialogHidden] = useState(true);
  const showAddRoleDialog = useCallback(
    () => setIsAddRoleDialogHidden(false),
    []
  );
  const hideAddRoleDialog = useCallback(
    () => setIsAddRoleDialogHidden(true),
    []
  );

  const filteredUserGroups = useMemo(() => {
    const userGroups =
      user.groups?.edges?.flatMap<UserGroupsListItem>((edge) => {
        if (edge?.node != null) {
          return [
            {
              ...edge.node,
              roles: {
                totalCount: edge.node.roles?.totalCount ?? 0,
                items:
                  edge.node.roles?.edges?.flatMap((edge) => {
                    return edge?.node == null ? [] : [edge.node];
                  }) ?? null,
              },
            },
          ];
        }
        return [];
      }) ?? [];
    if (isSearch) {
      return searchGroups(userGroups, searchKeyword);
    }

    return userGroups.slice(offset, offset + pageSize);
  }, [user.groups?.edges, isSearch, offset, searchKeyword]);

  const userGroups = useMemo(() => {
    return (
      user.groups?.edges?.flatMap((e) => {
        if (e?.node) {
          return [e.node];
        }
        return [];
      }) ?? []
    );
  }, [user.groups?.edges]);

  if (error != null) {
    return (
      <ShowError
        error={error}
        onRetry={() => {
          refetch().finally(() => {});
        }}
      />
    );
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
        <header className="flex flex-row items-center justify-between mb-5">
          <div className="max-w-[300px] min-w-0 flex-1 mr-2">
            <TextField
              size="2"
              type="search"
              iconStart={TextFieldIcon.MagnifyingGlass}
              placeholder={renderToString("search")}
              value={searchKeyword}
              onChange={onChangeSearchKeyword}
            />
          </div>
          <PrimaryButton
            size="2"
            text={
              <>
                <PlusIcon width="1rem" height="1rem" />
                <FormattedMessage id="UserDetailsScreen.groups.add" />
              </>
            }
            onClick={showAddRoleDialog}
          />
        </header>
        {userGroups.length === 0 ? (
          <RadixCallout.Root color="gray" size="2" variant="surface">
            <RadixCallout.Icon>
              <InfoCircledIcon width="1rem" height="1rem" />
            </RadixCallout.Icon>
            <RadixCallout.Text>
              <FormattedMessage id="UserDetailsScreen.groups.empty" />
            </RadixCallout.Text>
          </RadixCallout.Root>
        ) : (
          <UserGroupsList
            className="flex-1-0-auto min-h-[200px]"
            user={user}
            groups={filteredUserGroups}
            isSearch={isSearch}
            offset={offset}
            pageSize={pageSize}
            totalCount={userGroups.length}
            onChangeOffset={onChangeOffset}
          />
        )}
      </section>
      <AddUserGroupsDialog
        userID={user.id}
        userFormattedName={user.formattedName ?? null}
        userEndUserAccountID={user.endUserAccountID ?? null}
        userGroups={userGroups}
        isHidden={isAddRoleDialogHidden}
        onDismiss={hideAddRoleDialog}
      />
    </>
  );
}

export default UserDetailsScreenGroupListContainer;
