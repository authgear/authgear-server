import React, { useContext, useMemo } from "react";
import { useParams } from "react-router-dom";
import { useRoleQuery } from "./query/roleQuery";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { FormattedMessage, Context as MessageContext } from "../../intl";
import { RoleQueryNodeFragment } from "./query/roleQuery.generated";
import { usePivotNavigation } from "../../hook/usePivot";
import { BreadcrumbItem } from "../../NavBreadcrumb";
import { RoleAndGroupsLayout } from "../../RoleAndGroupsLayout";
import { OverflowTabs } from "../../components/v2/OverflowTabs/OverflowTabs";
import { RoleDetailsSettingsForm } from "../../components/roles-and-groups/form/RoleDetailsSettingsForm";
import RoleDetailsScreenGroupListContainer from "../../components/roles-and-groups/list/RoleDetailsScreenGroupListContainer";

const SETTINGS_KEY = "settings";
const GROUPS_KEY = "groups";

function RoleDetailsScreenLoaded(props: { role: RoleQueryNodeFragment }) {
  const { role } = props;
  const { renderToString } = useContext(MessageContext);

  const { selectedKey, onChangeKey } = usePivotNavigation([
    SETTINGS_KEY,
    GROUPS_KEY,
  ]);

  const tabs = useMemo(
    () => [
      {
        value: SETTINGS_KEY,
        label: renderToString("RoleDetailsScreen.tabs.settings"),
      },
      {
        value: GROUPS_KEY,
        label: renderToString("RoleDetailsScreen.tabs.groups"),
      },
    ],
    [renderToString]
  );

  const breadcrumbs = useMemo<BreadcrumbItem[]>(() => {
    return [
      {
        to: "~/user-management/roles",
        label: <FormattedMessage id="RolesScreen.title" />,
      },
      { to: ".", label: role.name ?? role.key },
    ];
  }, [role]);

  return (
    <RoleAndGroupsLayout headerBreadcrumbs={breadcrumbs}>
      <OverflowTabs
        className="mb-8"
        value={selectedKey}
        onValueChange={(value) => {
          onChangeKey(value as typeof SETTINGS_KEY | typeof GROUPS_KEY);
        }}
        tabs={tabs}
      />
      {selectedKey === GROUPS_KEY ? (
        <RoleDetailsScreenGroupListContainer role={role} />
      ) : (
        <RoleDetailsSettingsForm role={role} />
      )}
    </RoleAndGroupsLayout>
  );
}

const RoleDetailsScreen: React.VFC = function RoleDetailsScreen() {
  const { roleID } = useParams() as { roleID: string };
  const { role, loading, error, refetch } = useRoleQuery(roleID, {
    fetchPolicy: "network-only",
  });

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  if (loading) {
    return <ShowLoading />;
  }

  if (role == null) {
    return <ShowLoading />;
  }

  return <RoleDetailsScreenLoaded role={role} />;
};

export default RoleDetailsScreen;
