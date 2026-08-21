import React, { useCallback, useContext, useMemo, useState } from "react";
import { PlusIcon, Cross2Icon } from "@radix-ui/react-icons";
import { RoleQueryNodeFragment } from "../../../graphql/adminapi/query/roleQuery.generated";
import { FormattedMessage, Context as MessageContext } from "../../../intl";
import { useQuery } from "@apollo/client";
import ShowError from "../../../ShowError";
import ShowLoading from "../../../ShowLoading";
import { PrimaryButton } from "../../v2/Button/PrimaryButton/PrimaryButton";
import { TextField, TextFieldIcon } from "../../v2/TextField/TextField";
import {
  GroupsListQueryDocument,
  GroupsListQueryQuery,
  GroupsListQueryQueryVariables,
} from "../../../graphql/adminapi/query/groupsListQuery.generated";
import { GroupsEmptyView } from "../empty-view/GroupsEmptyView";
import { RoleGroupsList, RoleGroupsListItem } from "../list/RoleGroupsList";
import { AddRoleGroupsDialog } from "../dialog/AddRoleGroupsDialog";
import { searchGroups } from "../../../model/group";

interface RoleDetailsScreenGroupListContainerProps {
  role: RoleQueryNodeFragment;
}

const RoleDetailsScreenGroupListContainer: React.VFC<
  RoleDetailsScreenGroupListContainerProps
> = ({ role }) => {
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
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setSearchKeyword(e.target.value);
    },
    []
  );
  const onClearSearchKeyword = useCallback(() => {
    setSearchKeyword("");
  }, []);

  const [isAddGroupDialogHidden, setIsAddGroupDialogHidden] = useState(true);
  const showAddGroupDialog = useCallback(
    () => setIsAddGroupDialogHidden(false),
    []
  );
  const hideAddGroupDialog = useCallback(
    () => setIsAddGroupDialogHidden(true),
    []
  );

  const filteredRoleGroups = useMemo(() => {
    const roleGroups =
      role.groups?.edges?.flatMap<RoleGroupsListItem>((edge) => {
        if (edge?.node != null) {
          return [edge.node];
        }
        return [];
      }) ?? [];
    return searchGroups(roleGroups, searchKeyword);
  }, [role.groups?.edges, searchKeyword]);

  const roleGroups = useMemo(() => {
    return (
      role.groups?.edges?.flatMap((e) => {
        if (e?.node) {
          return [e.node];
        }
        return [];
      }) ?? []
    );
  }, [role.groups?.edges]);

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
      <section className="flex-1 flex flex-col">
        <header className="flex flex-row items-center justify-between mb-8 gap-2">
          <div className="max-w-[300px] min-w-0 flex-1">
            <TextField
              size="2"
              type="search"
              value={searchKeyword}
              placeholder={renderToString("search")}
              iconStart={TextFieldIcon.MagnifyingGlass}
              onChange={onChangeSearchKeyword}
              suffixPlain={true}
              suffix={
                searchKeyword !== "" ? (
                  <button
                    type="button"
                    className="inline-flex items-center justify-center border-0 bg-transparent p-0 cursor-pointer"
                    aria-label={renderToString(
                      "APIResourcesScreen.clear-search"
                    )}
                    onClick={onClearSearchKeyword}
                  >
                    <Cross2Icon width="0.875rem" height="0.875rem" />
                  </button>
                ) : undefined
              }
            />
          </div>
          <PrimaryButton
            size="2"
            onClick={showAddGroupDialog}
            text={
              <span className="inline-flex items-center gap-1">
                <PlusIcon width="1rem" height="1rem" />
                <FormattedMessage id="RoleDetailsScreen.groups.add" />
              </span>
            }
          />
        </header>
        <RoleGroupsList
          className="flex-1 min-h-0"
          role={role}
          groups={filteredRoleGroups}
        />
      </section>
      <AddRoleGroupsDialog
        roleID={role.id}
        roleKey={role.key}
        roleName={role.name ?? null}
        roleGroups={roleGroups}
        isHidden={isAddGroupDialogHidden}
        onDismiss={hideAddGroupDialog}
      />
    </>
  );
};

export default RoleDetailsScreenGroupListContainer;
