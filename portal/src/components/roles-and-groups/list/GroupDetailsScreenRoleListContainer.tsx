import React, { useCallback, useContext, useMemo, useState } from "react";
import { PlusIcon, Cross2Icon } from "@radix-ui/react-icons";
import { GroupQueryNodeFragment } from "../../../graphql/adminapi/query/groupQuery.generated";
import { FormattedMessage, Context as MessageContext } from "../../../intl";
import { useQuery } from "@apollo/client";
import ShowError from "../../../ShowError";
import ShowLoading from "../../../ShowLoading";
import { Role } from "../../../graphql/adminapi/globalTypes.generated";
import { PrimaryButton } from "../../v2/Button/PrimaryButton/PrimaryButton";
import { TextField, TextFieldIcon } from "../../v2/TextField/TextField";
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
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setSearchKeyword(e.target.value);
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
    // eslint-disable-next-line @typescript-eslint/strict-void-return
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
            onClick={showAddRoleDialog}
            text={
              <span className="inline-flex items-center gap-1">
                <PlusIcon width="1rem" height="1rem" />
                <FormattedMessage id="GroupDetailsScreen.roles.add" />
              </span>
            }
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
