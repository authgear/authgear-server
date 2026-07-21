import React from "react";
import { Navigate, Outlet, useParams } from "react-router-dom";
import { useUserQuery } from "./graphql/adminapi/query/userQuery";
import ShowLoading from "./ShowLoading";
import ShowError from "./ShowError";

// RequireUser is a route-level guard for routes under
// /project/:appID/user-management/users/:userID. When the user does not exist
// in the current project (e.g. after switching projects with a stale URL), it
// redirects to the user list; otherwise it renders the matched child route via
// <Outlet>. The user query is cached, so child screens that query the same
// user do not refetch.
const RequireUser: React.VFC = function RequireUser() {
  const { appID, userID } = useParams() as { appID: string; userID: string };
  const { user, loading, error, refetch } = useUserQuery(userID);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    // eslint-disable-next-line @typescript-eslint/strict-void-return
    return <ShowError error={error} onRetry={refetch} />;
  }

  if (user == null) {
    return (
      <Navigate to={`/project/${appID}/user-management/users`} replace={true} />
    );
  }

  return <Outlet />;
};

export default RequireUser;
