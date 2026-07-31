import React, { useContext, useEffect, useMemo } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useGroupQuery } from "./query/groupQuery";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { FormattedMessage, Context as MessageContext } from "../../intl";
import { GroupQueryNodeFragment } from "./query/groupQuery.generated";
import { usePivotNavigation } from "../../hook/usePivot";
import { BreadcrumbItem } from "../../NavBreadcrumb";
import { RoleAndGroupsLayout } from "../../RoleAndGroupsLayout";
import { OverflowTabs } from "../../components/v2/OverflowTabs/OverflowTabs";
import { GroupDetailsSettingsForm } from "../../components/roles-and-groups/form/GroupDetailsSettingsForm";
import GroupDetailsScreenRoleListContainer from "../../components/roles-and-groups/list/GroupDetailsScreenRoleListContainer";

const SETTINGS_KEY = "settings";
const ROLES_KEY = "roles";

function GroupDetailsScreenLoaded(props: { group: GroupQueryNodeFragment }) {
  const { group } = props;
  const { renderToString } = useContext(MessageContext);

  const { selectedKey, onChangeKey } = usePivotNavigation([
    SETTINGS_KEY,
    ROLES_KEY,
  ]);

  const tabs = useMemo(
    () => [
      {
        value: SETTINGS_KEY,
        label: renderToString("GroupDetailsScreen.tabs.settings"),
      },
      {
        value: ROLES_KEY,
        label: renderToString("GroupDetailsScreen.tabs.roles"),
      },
    ],
    [renderToString]
  );

  const breadcrumbs = useMemo<BreadcrumbItem[]>(() => {
    return [
      {
        to: "~/user-management/groups",
        label: <FormattedMessage id="GroupsScreen.title" />,
      },
      { to: ".", label: group.name ?? group.key },
    ];
  }, [group]);

  return (
    <RoleAndGroupsLayout headerBreadcrumbs={breadcrumbs}>
      <OverflowTabs
        className="mb-8"
        value={selectedKey}
        onValueChange={(value) => {
          onChangeKey(value as typeof SETTINGS_KEY | typeof ROLES_KEY);
        }}
        tabs={tabs}
      />
      {selectedKey === ROLES_KEY ? (
        <GroupDetailsScreenRoleListContainer group={group} />
      ) : (
        <GroupDetailsSettingsForm group={group} />
      )}
    </RoleAndGroupsLayout>
  );
}

const GroupDetailsScreen: React.VFC = function GroupDetailsScreen() {
  const { appID, groupID } = useParams() as { appID: string; groupID: string };
  const navigate = useNavigate();
  const { group, loading, error, refetch } = useGroupQuery(groupID, {
    fetchPolicy: "network-only",
  });

  useEffect(() => {
    if (loading) {
      return;
    }
    if (error != null) {
      return;
    }
    if (group != null) {
      return;
    }
    navigate(`/project/${appID}/user-management/groups`, { replace: true });
  }, [appID, error, group, loading, navigate]);

  if (error != null) {
    // eslint-disable-next-line @typescript-eslint/strict-void-return
    return <ShowError error={error} onRetry={refetch} />;
  }

  if (loading) {
    return <ShowLoading />;
  }

  if (group == null) {
    return <ShowLoading />;
  }

  return <GroupDetailsScreenLoaded group={group} />;
};

export default GroupDetailsScreen;
