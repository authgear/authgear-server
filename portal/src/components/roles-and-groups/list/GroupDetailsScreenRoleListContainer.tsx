import React, { useCallback, useContext, useMemo, useState } from "react";
import { GroupQueryNodeFragment } from "../../../graphql/adminapi/query/groupQuery.generated";
import {
  FormattedMessage,
  Context as MessageContext,
} from "../../../intl";
import { useQuery } from "@apollo/client";
import { SearchBox } from "@fluentui/react";
import ShowError from "../../../ShowError";
import ShowLoading from "../../../ShowLoading";
import { Role } from "../../../graphql/adminapi/globalTypes.generated";
import PrimaryButton from "../../../PrimaryButton";
import {
  RolesListQueryDocument,
  RolesListQueryQuery,
  RolesListQueryQueryVariables,
} from "../../../graphql/adminapi/query/rolesListQuery.generated";
import { RolesEmptyView } from "../empty-view/RolesEmptyView";
import { GroupRolesList } from "../list/GroupRolesList";
import { AddGroupRolesDialog } from "../dialog/AddGroupRolesDialog";
import { searchRoles } from "../../../model/role";

export interface GroupRolesListItem extends Pick<Role, "id" | "name" | "key"> {}

interface GroupDetailsScreenRoleListContainerProps {
  group: GroupQueryNodeFragment;
}

const GroupDetailsScreenRoleListContainer: React.VFC<
  GroupDetailsScreenRoleListContainerProps
> = ({ group }) => {
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

  const [isAddRoleDialogHidden, setIsAddRoleDialogHidden] = useState(true);
  const showAddRoleDialog = useCallback(
    () => setIsAddRoleDialogHidden(false),
    []
  );
  const hideAddRoleDialog = useCallback(
    () => setIsAddRoleDialogHidden(true),
    []
  );

  const filteredGroupRoles = useMemo(() => {
    const groupRoles =
      group.roles?.edges?.flatMap<GroupRolesListItem>((edge) => {
        if (edge?.node != null) {
          return [edge.node];
        }
        return [];
      }) ?? [];
    return searchRoles(groupRoles, searchKeyword);
  }, [group.roles?.edges, searchKeyword]);

  const groupRoles = useMemo(() => {
    return (
      group.roles?.edges?.flatMap((e) => {
        if (e?.node) {
          return [e.node];
        }
        return [];
      }) ?? []
    );
  }, [group.roles?.edges]);

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
      <section className="flex-1 flex flex-col">
        <header className="flex flex-row items-center justify-between mb-8">
          <SearchBox
            className="max-w-[300px] min-w-0 flex-1 mr-2"
            placeholder={renderToString("search")}
            value={searchKeyword}
            onChange={onChangeSearchKeyword}
            onClear={onClearSearchKeyword}
          />
          <PrimaryButton
            text={<FormattedMessage id="GroupDetailsScreen.roles.add" />}
            onClick={showAddRoleDialog}
          />
        </header>
        <GroupRolesList
          className="flex-1 min-h-0"
          group={group}
          roles={filteredGroupRoles}
        />
      </section>
      <AddGroupRolesDialog
        groupID={group.id}
        groupKey={group.key}
        groupName={group.name ?? null}
        groupRoles={groupRoles}
        isHidden={isAddRoleDialogHidden}
        onDismiss={hideAddRoleDialog}
      />
    </>
  );
};

export default GroupDetailsScreenRoleListContainer;
