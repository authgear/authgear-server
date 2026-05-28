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
import { PivotItem } from "@fluentui/react";
import { AGPivot } from "../../components/common/AGPivot";
import { GroupDetailsSettingsForm } from "../../components/roles-and-groups/form/GroupDetailsSettingsForm";
import GroupDetailsScreenRoleListContainer from "../../components/roles-and-groups/list/GroupDetailsScreenRoleListContainer";

const SETTINGS_KEY = "settings";
const ROLES_KEY = "roles";

function GroupDetailsScreenLoaded(props: { group: GroupQueryNodeFragment }) {
  const { group } = props;
  const { renderToString } = useContext(MessageContext);

  const { selectedKey, onLinkClick } = usePivotNavigation([
    SETTINGS_KEY,
    ROLES_KEY,
  ]);

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
      <AGPivot
        overflowBehavior="menu"
        selectedKey={selectedKey}
        onLinkClick={onLinkClick}
        className="mb-8"
      >
        <PivotItem
          itemKey={SETTINGS_KEY}
          headerText={renderToString("GroupDetailsScreen.tabs.settings")}
        />
        <PivotItem
          itemKey={ROLES_KEY}
          headerText={renderToString("GroupDetailsScreen.tabs.roles")}
        />
      </AGPivot>
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
