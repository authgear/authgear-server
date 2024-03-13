import React, { useCallback, useContext, useMemo, useState } from "react";
import { GroupQueryNodeFragment } from "../../graphql/adminapi/query/groupQuery.generated";
import {
  FormattedMessage,
  Context as MessageContext,
} from "@oursky/react-messageformat";
import { useQuery } from "@apollo/client";
import { SearchBox } from "@fluentui/react";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { Role } from "../../graphql/adminapi/globalTypes.generated";
import PrimaryButton from "../../PrimaryButton";
import {
  RolesListQueryDocument,
  RolesListQueryQuery,
  RolesListQueryQueryVariables,
} from "../../graphql/adminapi/query/rolesListQuery.generated";
import { RolesEmptyView } from "./RolesEmptyView";
import { GroupRolesList } from "./GroupRolesList";
import { searchRolesAndGroups } from "../../util/rolesAndGroups";

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

  const filteredGroupRoles = useMemo(() => {
    const groupRoles =
      group.roles?.edges?.flatMap<GroupRolesListItem>((edge) => {
        if (edge?.node != null) {
          return [edge.node];
        }
        return [];
      }) ?? [];
    return searchRolesAndGroups(groupRoles, searchKeyword);
  }, [group.roles?.edges, searchKeyword]);

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
          />
        </header>
        <GroupRolesList
          className="flex-1 min-h-0"
          group={group}
          roles={filteredGroupRoles}
        />
      </section>
    </>
  );
};

export default GroupDetailsScreenRoleListContainer;
