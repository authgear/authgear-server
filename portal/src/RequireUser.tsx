import React from "react";
import { useParams } from "react-router-dom";
import { useUserQuery } from "./graphql/adminapi/query/userQuery";
import { EntityRouteGuard } from "./EntityRouteGuard";

// Route-level guard for /project/:appID/user-management/users/:userID; see
// EntityRouteGuard. The user query is cached, so child screens that query the
// same user do not refetch.
const RequireUser: React.VFC = function RequireUser() {
  const { appID, userID } = useParams() as { appID: string; userID: string };
  const { user, loading, error, refetch } = useUserQuery(userID);

  return (
    <EntityRouteGuard
      loading={loading}
      error={error}
      refetch={refetch}
      entity={user}
      redirectTo={`/project/${appID}/user-management/users`}
    />
  );
};

export default RequireUser;
