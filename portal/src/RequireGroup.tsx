import React from "react";
import { useParams } from "react-router-dom";
import { useGroupQuery } from "./graphql/adminapi/query/groupQuery";
import { EntityRouteGuard } from "./EntityRouteGuard";

// Route-level guard for /project/:appID/user-management/groups/:groupID; see
// EntityRouteGuard.
const RequireGroup: React.VFC = function RequireGroup() {
  const { appID, groupID } = useParams() as { appID: string; groupID: string };
  const { group, loading, error, refetch } = useGroupQuery(groupID);

  return (
    <EntityRouteGuard
      loading={loading}
      error={error}
      refetch={refetch}
      entity={group}
      redirectTo={`/project/${appID}/user-management/groups`}
    />
  );
};

export default RequireGroup;
