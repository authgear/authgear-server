import React from "react";
import { useParams } from "react-router-dom";
import { useRoleQuery } from "./graphql/adminapi/query/roleQuery";
import { EntityRouteGuard } from "./EntityRouteGuard";

// Route-level guard for /project/:appID/user-management/roles/:roleID; see
// EntityRouteGuard.
const RequireRole: React.VFC = function RequireRole() {
  const { appID, roleID } = useParams() as { appID: string; roleID: string };
  const { role, loading, error, refetch } = useRoleQuery(roleID);

  return (
    <EntityRouteGuard
      loading={loading}
      error={error}
      refetch={refetch}
      entity={role}
      redirectTo={`/project/${appID}/user-management/roles`}
    />
  );
};

export default RequireRole;
