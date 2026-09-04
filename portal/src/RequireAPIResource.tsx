import React from "react";
import { useParams } from "react-router-dom";
import { useResourceQueryQuery } from "./graphql/adminapi/query/resourceQuery.generated";
import { EntityRouteGuard } from "./EntityRouteGuard";

// Route-level guard for /project/:appID/api-resources/:resourceID; see
// EntityRouteGuard. The resource query is cached, so child screens that query
// the same resource do not refetch.
const RequireAPIResource: React.VFC = function RequireAPIResource() {
  const { appID, resourceID } = useParams() as {
    appID: string;
    resourceID: string;
  };
  const { data, loading, error, refetch } = useResourceQueryQuery({
    variables: { id: resourceID },
  });

  const resource = data?.node?.__typename === "Resource" ? data.node : null;

  return (
    <EntityRouteGuard
      loading={loading}
      error={error}
      refetch={refetch}
      entity={resource}
      redirectTo={`/project/${appID}/api-resources`}
    />
  );
};

export default RequireAPIResource;
